package workflow

import (
	"context"
	"fmt"
	"time"
)

// LogFunc 日志回调函数
type LogFunc func(log *ExecutionLog)

// SecretStore 全局密钥存储接口
type SecretStore interface {
	GetAll(ctx context.Context) (map[string]any, error)
}

// DAG 工作流执行引擎
type Engine struct {
	registry    *ExecutorRegistry
	secretStore SecretStore
}

func NewEngine(registry *ExecutorRegistry) *Engine {
	en := &Engine{
		registry: registry,
	}
	return en
}

func (e *Engine) SetSecretStore(store SecretStore) {
	e.secretStore = store
}

func (e *Engine) Run(ctx context.Context, wf *Workflow, input map[string]any, eventBus *EventBus, logFn LogFunc) (map[string]any, map[string]NodeState, error) {
	graph := buildGraph(wf)
	nodeStates := make(map[string]NodeState)
	nodeOutputs := make(map[string]map[string]any)

	// 存储工作流输入，供模板用 {{input.xxx}} 引用
	nodeOutputs["__input__"] = input

	// 存储工作流变量，供模板用 {{var.xxx}} 引用
	if wf.Variables != nil {
		nodeOutputs["var"] = wf.Variables
	}
	// 存储全局密钥，供模板用 {{secret.xxx}} 引用
	if e.secretStore != nil {
		if secrets, err := e.secretStore.GetAll(ctx); err == nil {
			nodeOutputs["secret"] = secrets
		}
	}

	// 从开始节点开始
	startID := ""
	for _, n := range wf.Nodes {
		if n.Type == NodeStart {
			startID = n.ID
			break
		}
	}
	if startID == "" {
		return nil, nil, fmt.Errorf("workflow has no start node")
	}

	err := e.executeNode(ctx, graph, startID, input, nodeStates, nodeOutputs, eventBus, logFn)

	if err != nil {
		if eventBus != nil {
			eventBus.Emit(Event{Type: EventFlowFailed, Error: err.Error()})
			eventBus.Close()
		}
		return nil, nodeStates, err
	}

	// 找到 end 节点 作为输出结果
	var output map[string]any
	for _, n := range wf.Nodes {
		if n.Type == NodeEnd {
			if o, ok := nodeOutputs[n.ID]; ok {
				output = o
			}
			break
		}
	}
	if eventBus != nil {
		eventBus.Emit(Event{Type: EventFlowCompleted, Output: output})
		eventBus.Close()
	}

	return output, nodeStates, nil
}

func (e *Engine) executeNode(
	ctx context.Context,
	graph *dagGraph,
	nodeID string,
	input map[string]any,
	nodeStates map[string]NodeState,
	nodeOutputs map[string]map[string]any,
	eventBus *EventBus,
	logFn LogFunc,
) error {
	if _, done := nodeStates[nodeID]; done {
		return nil
	}

	node, ok := graph.nodes[nodeID]
	if !ok {
		return fmt.Errorf("node not found %s", nodeID)
	}

	// 执行当前节点
	executor, ok := e.registry.Get(node.Type)
	if !ok {
		return fmt.Errorf("node type %v executor not found", node.Type)
	}

	// 渲染模板：input + node config
	renderedInput := RenderMap(input, nodeOutputs, nodeOutputs["__input__"])
	node = renderNodeConfig(node, nodeOutputs, nodeOutputs["__input__"])
	start := time.Now()
	nodeStates[nodeID] = NodeState{NodeID: nodeID, Status: string(ExecRunning)}

	// 节点开始事件
	if eventBus != nil {
		eventBus.Emit(Event{
			Type: EventNodeStarted, NodeID: nodeID,
			NodeName: node.Name, NodeType: node.Type,
		})
	}

	// 重试配置
	maxRetries := 0
	retryInterval := time.Second
	if node.Retry != nil {
		maxRetries = node.Retry.MaxRetries
		if node.Retry.Interval > 0 {
			retryInterval = time.Duration(node.Retry.Interval) * time.Second
		}
	}

	// 超时配置
	timeout := node.Timeout
	if timeout <= 0 {
		timeout = 60
	}

	var output map[string]any
	var err error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		attemptStart := time.Now()
		output, err = executor.Execute(execCtx, node, renderedInput)
		cancel()
		attemptDuration := time.Since(attemptStart).Milliseconds()

		// 写执行日志
		if logFn != nil {
			l := &ExecutionLog{
				NodeID:   nodeID,
				NodeName: node.Name,
				NodeType: node.Type,
				Attempt:  attempt + 1,
				Input:    renderedInput,
				Duration: attemptDuration,
			}
			if err != nil {
				l.Status = string(ExecFailed)
				l.Error = err.Error()
			} else {
				l.Status = string(ExecCompleted)
				l.Output = output
			}
			logFn(l)
		}

		if err == nil {
			break
		}
		if attempt < maxRetries {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryInterval):
			}
		}
	}

	duration := time.Since(start).Milliseconds()
	if err != nil {
		nodeStates[nodeID] = NodeState{
			NodeID:   nodeID,
			Status:   string(ExecFailed),
			Input:    input,
			Error:    err.Error(),
			Duration: duration,
		}
		if eventBus != nil {
			eventBus.Emit(Event{
				Type: EventNodeFailed, NodeID: nodeID,
				NodeName: node.Name, NodeType: node.Type,
				Error: err.Error(), Duration: duration,
			})
		}
		return fmt.Errorf("node %s(%s) failed %w", node.Name, nodeID, err)
	}

	nodeStates[nodeID] = NodeState{
		NodeID:   nodeID,
		Status:   string(ExecCompleted),
		Input:    input,
		Output:   output,
		Duration: duration,
	}
	nodeOutputs[nodeID] = output

	// 节点完成事件
	if eventBus != nil {
		eventBus.Emit(Event{
			Type: EventNodeCompleted, NodeID: nodeID,
			NodeName: node.Name, NodeType: node.Type,
			Output: output, Duration: duration,
		})
	}
	// 继续执行下游节点
	nextIDs := e.resolveNext(node, output, graph)

	for _, nextID := range nextIDs {
		if err := e.executeNode(ctx, graph, nextID, output, nodeStates, nodeOutputs, eventBus, logFn); err != nil {
			return err
		}
	}

	return nil
}

func (e *Engine) resolveNext(node *Node, output map[string]any, graph *dagGraph) []string {
	edges := graph.outEdges[node.ID]

	// 条件节点 根据 branch 值选择对应的边
	if node.Type == NodeCondition {
		branch, _ := output["branch"].(string)
		for _, edge := range edges {
			if edge.Condition == branch {
				return []string{edge.Target}
			}
		}
		return nil
	}

	// 普通节点 所有下游
	var next []string
	for _, edge := range edges {
		next = append(next, edge.Target)
	}
	return next
}
