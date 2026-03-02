package workflow

import (
	"context"
	"fmt"
	"time"
)

// DAG 工作流执行引擎
type Engine struct {
	registry *ExecutorRegistry
}

func NewEngine(registry *ExecutorRegistry) *Engine {
	en := &Engine{
		registry: registry,
	}
	return en
}

func (e *Engine) Run(ctx context.Context, wf *Workflow, input map[string]any, eventBus *EventBus) (map[string]any, map[string]NodeState, error) {
	graph := buildGraph(wf)
	nodeStates := make(map[string]NodeState)
	nodeOutputs := make(map[string]map[string]any)

	// 存储工作流输入，供模板用 {{input.xxx}} 引用
	nodeOutputs["__input__"] = input

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

	err := e.executeNode(ctx, graph, startID, input, nodeStates, nodeOutputs, eventBus)

	if err != nil {
		// 执行完发结束事件
		eventBus.Emit(Event{Type: EventFlowFailed, Error: err.Error()})
		eventBus.Close()
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
	eventBus.Emit(Event{Type: EventFlowCompleted, Output: output})
	eventBus.Close()

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

	// 渲染模板
	renderedInput := RenderMap(input, nodeOutputs, nodeOutputs["__input__"])
	start := time.Now()
	nodeStates[nodeID] = NodeState{NodeID: nodeID, Status: string(ExecRunning)}

	// 节点开始事件
	if eventBus != nil {
		eventBus.Emit(Event{
			Type: EventNodeStarted, NodeID: nodeID,
			NodeName: node.Name, NodeType: node.Type,
		})
	}
	output, err := executor.Execute(ctx, node, renderedInput)
	duration := time.Since(start).Milliseconds()
	if err != nil {
		nodeStates[nodeID] = NodeState{
			NodeID:   nodeID,
			Status:   string(ExecFailed),
			Input:    input,
			Error:    err.Error(),
			Duration: duration,
		}
		return fmt.Errorf("node %s(%s) failed: %w", node.Name, nodeID, err)
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
		if err := e.executeNode(ctx, graph, nextID, output, nodeStates, nodeOutputs, eventBus); err != nil {
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
