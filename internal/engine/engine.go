package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/2comjie/mcpflow/internal/model"
)

// LogFunc 执行日志回调
type LogFunc func(log *model.ExecutionLog)

// Engine 工作流执行引擎
type Engine struct {
	registry *ExecutorRegistry
}

func NewEngine(registry *ExecutorRegistry) *Engine {
	return &Engine{registry: registry}
}

// RunResult 执行结果
type RunResult struct {
	Output     map[string]any
	NodeStates model.NodeStates
}

// Run 执行工作流
func (e *Engine) Run(ctx context.Context, wf *model.Workflow, input map[string]any, logFn LogFunc) (*RunResult, error) {
	graph, err := buildGraph(wf.Nodes, wf.Edges)
	if err != nil {
		return nil, fmt.Errorf("build graph: %w", err)
	}

	startNode, err := graph.findStartNode()
	if err != nil {
		return nil, err
	}

	// 节点输出缓存：nodeID -> output
	outputs := make(map[string]map[string]any)
	nodeStates := make(model.NodeStates)

	// 模板渲染数据：input 字段保留完整引用，同时平铺到顶层方便使用
	templateData := map[string]any{
		"input": input,
	}
	for k, v := range input {
		templateData[k] = v
	}

	// 从 start 节点递归执行
	if err := e.executeNode(ctx, graph, startNode, input, outputs, nodeStates, templateData, logFn, 1); err != nil {
		return nil, err
	}

	// 找到 end 节点的输出作为工作流输出
	var finalOutput map[string]any
	for id, node := range graph.nodes {
		if node.Type == model.NodeEnd {
			if out, ok := outputs[id]; ok {
				finalOutput = out
				break
			}
		}
	}
	if finalOutput == nil {
		finalOutput = map[string]any{}
	}

	return &RunResult{
		Output:     finalOutput,
		NodeStates: nodeStates,
	}, nil
}

func (e *Engine) executeNode(
	ctx context.Context,
	graph *dagGraph,
	node *model.Node,
	input map[string]any,
	outputs map[string]map[string]any,
	nodeStates model.NodeStates,
	templateData map[string]any,
	logFn LogFunc,
	attempt int,
) error {
	// 已执行过则跳过
	if _, done := outputs[node.ID]; done {
		return nil
	}

	// 设置超时
	nodeCtx := ctx
	timeout := node.Timeout
	if timeout <= 0 {
		timeout = 60
	}
	nodeCtx, cancel := context.WithTimeout(nodeCtx, time.Duration(timeout)*time.Second)
	defer cancel()

	// 渲染节点配置中的模板
	renderedNode := *node
	renderedNode.Config = RenderNodeConfig(node, templateData)

	// 获取执行器
	executor, err := e.registry.Get(node.Type)
	if err != nil {
		e.recordState(nodeStates, node, "failed", input, nil, err.Error(), 0)
		return err
	}

	// 执行节点
	start := time.Now()
	output, err := executor.Execute(nodeCtx, &renderedNode, input)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		// 重试逻辑
		if node.Retry != nil && attempt <= node.Retry.Max {
			time.Sleep(time.Duration(node.Retry.Interval) * time.Second)
			return e.executeNode(ctx, graph, node, input, outputs, nodeStates, templateData, logFn, attempt+1)
		}
		e.recordState(nodeStates, node, "failed", input, nil, err.Error(), duration)
		e.writeLog(logFn, node, attempt, "failed", input, nil, err.Error(), duration)
		return fmt.Errorf("node %s (%s) failed: %w", node.ID, node.Name, err)
	}

	// 记录成功
	outputs[node.ID] = output
	templateData[node.ID] = output
	e.recordState(nodeStates, node, "completed", input, output, "", duration)
	e.writeLog(logFn, node, attempt, "completed", input, output, "", duration)

	// 条件节点：根据 branch 选择下游
	branch := ""
	if node.Type == model.NodeCondition {
		if b, ok := output["branch"].(string); ok {
			branch = b
		}
	}

	// 递归执行下游节点
	nextNodes := graph.getNextNodes(node.ID, branch)
	for _, next := range nextNodes {
		if err := e.executeNode(ctx, graph, next, output, outputs, nodeStates, templateData, logFn, 1); err != nil {
			return err
		}
	}

	return nil
}

func (e *Engine) recordState(states model.NodeStates, node *model.Node, status string, input, output map[string]any, errMsg string, duration int64) {
	states[node.ID] = model.NodeState{
		NodeID:   node.ID,
		Status:   status,
		Input:    input,
		Output:   output,
		Error:    errMsg,
		Duration: duration,
	}
}

func (e *Engine) writeLog(logFn LogFunc, node *model.Node, attempt int, status string, input, output map[string]any, errMsg string, duration int64) {
	if logFn == nil {
		return
	}
	logFn(&model.ExecutionLog{
		NodeID:   node.ID,
		NodeName: node.Name,
		NodeType: node.Type,
		Attempt:  attempt,
		Status:   status,
		Input:    input,
		Output:   output,
		Error:    errMsg,
		Duration: duration,
	})
}
