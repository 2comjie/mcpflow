package engine

import (
	"fmt"
	"time"

	"github.com/2comjie/mcpflow/internal/model"
	"github.com/2comjie/mcpflow/internal/store"
)

type Engine struct {
	store *store.Store
}

func New(s *store.Store) *Engine {
	return &Engine{store: s}
}

// ExecuteWorkflow 执行工作流：构建 DAG → 拓扑排序 → 逐节点执行
func (e *Engine) ExecuteWorkflow(wf *model.Workflow, input map[string]any) (*model.Execution, error) {
	exec := &model.Execution{
		WorkflowID: wf.ID,
		Status:     model.ExecRunning,
		Input:      input,
		NodeStates: make(map[string]NodeState),
	}
	now := time.Now()
	exec.StartedAt = &now
	if err := e.store.CreateExecution(exec); err != nil {
		return nil, fmt.Errorf("create execution: %w", err)
	}

	order, err := topoSort(wf.Nodes, wf.Edges)
	if err != nil {
		return e.failExecution(exec, err)
	}

	nodeMap := make(map[string]*model.Node)
	for i := range wf.Nodes {
		nodeMap[wf.Nodes[i].ID] = &wf.Nodes[i]
	}
	edgeMap := buildEdgeMap(wf.Edges)

	// ctx 存储节点间传递的数据，key 为 node_id
	ctx := &WorkflowContext{
		Input:      input,
		NodeOutput: make(map[string]any),
		EdgeMap:    edgeMap,
	}

	for _, nodeID := range order {
		node := nodeMap[nodeID]
		if node == nil {
			continue
		}

		start := time.Now()
		output, err := e.executeNode(node, ctx)
		duration := time.Since(start).Milliseconds()

		state := NodeState{
			NodeID:   node.ID,
			Status:   "completed",
			Output:   output,
			Duration: duration,
		}
		if err != nil {
			state.Status = "failed"
			state.Error = err.Error()
		}
		exec.NodeStates[node.ID] = state

		// 记录执行日志
		log := &model.ExecutionLog{
			ExecutionID: exec.ID,
			NodeID:      node.ID,
			NodeName:    node.Name,
			NodeType:    string(node.Type),
			Status:      state.Status,
			Output:      toMapAny(output),
			Error:       state.Error,
			Duration:    duration,
		}
		if steps, ok := output.(AgentResult); ok {
			log.AgentSteps = steps.Steps
			log.Output = map[string]any{"content": steps.Content}
		}
		_ = e.store.CreateExecutionLog(log)

		if err != nil {
			return e.failExecution(exec, err)
		}

		// 条件节点：决定走哪条边
		if node.Type == model.NodeCondition {
			ctx.NodeOutput[node.ID] = output
			condResult, _ := output.(bool)
			ctx.ConditionResults = append(ctx.ConditionResults, ConditionResult{
				NodeID: node.ID,
				Result: condResult,
			})
		} else {
			ctx.NodeOutput[node.ID] = output
		}
	}

	// 完成
	finished := time.Now()
	exec.FinishedAt = &finished
	exec.Status = model.ExecCompleted
	exec.Output = toMapAny(ctx.NodeOutput)
	_ = e.store.UpdateExecution(exec.ID, map[string]any{
		"status":      exec.Status,
		"output":      exec.Output,
		"node_states": exec.NodeStates,
		"finished_at": exec.FinishedAt,
	})
	return exec, nil
}

func (e *Engine) failExecution(exec *model.Execution, err error) (*model.Execution, error) {
	finished := time.Now()
	exec.FinishedAt = &finished
	exec.Status = model.ExecFailed
	exec.Error = err.Error()
	_ = e.store.UpdateExecution(exec.ID, map[string]any{
		"status":      exec.Status,
		"error":       exec.Error,
		"node_states": exec.NodeStates,
		"finished_at": exec.FinishedAt,
	})
	return exec, err
}

func (e *Engine) executeNode(node *model.Node, ctx *WorkflowContext) (any, error) {
	switch node.Type {
	case model.NodeStart:
		return executeStart(ctx)
	case model.NodeEnd:
		return executeEnd(ctx)
	case model.NodeLLM:
		return executeLLM(node.Config.LLM, ctx)
	case model.NodeAgent:
		return executeAgent(node.Config.Agent, ctx)
	case model.NodeCondition:
		return executeCondition(node.Config.Condition, ctx)
	case model.NodeCode:
		return executeCode(node.Config.Code, ctx)
	case model.NodeHTTP:
		return executeHTTP(node.Config.HTTP, ctx)
	case model.NodeEmail:
		return executeEmail(node.Config.Email, ctx)
	default:
		return nil, fmt.Errorf("unknown node type: %s", node.Type)
	}
}

// WorkflowContext 工作流执行上下文
type WorkflowContext struct {
	Input            map[string]any
	NodeOutput       map[string]any
	EdgeMap          map[string][]EdgeInfo
	ConditionResults []ConditionResult
}

type EdgeInfo struct {
	TargetID  string
	Condition string
}

type ConditionResult struct {
	NodeID string
	Result bool
}

type NodeState = model.NodeState

func buildEdgeMap(edges []model.Edge) map[string][]EdgeInfo {
	m := make(map[string][]EdgeInfo)
	for _, e := range edges {
		m[e.Source] = append(m[e.Source], EdgeInfo{
			TargetID:  e.Target,
			Condition: e.Condition,
		})
	}
	return m
}

// topoSort 对节点进行拓扑排序
func topoSort(nodes []model.Node, edges []model.Edge) ([]string, error) {
	inDegree := make(map[string]int)
	adj := make(map[string][]string)

	for _, n := range nodes {
		inDegree[n.ID] = 0
	}
	for _, e := range edges {
		adj[e.Source] = append(adj[e.Source], e.Target)
		inDegree[e.Target]++
	}

	var queue []string
	for _, n := range nodes {
		if inDegree[n.ID] == 0 {
			queue = append(queue, n.ID)
		}
	}

	var order []string
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		order = append(order, curr)
		for _, next := range adj[curr] {
			inDegree[next]--
			if inDegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}

	if len(order) != len(nodes) {
		return nil, fmt.Errorf("workflow has cycle")
	}
	return order, nil
}

// getPreviousOutput 获取上一个节点的输出（用于节点间数据传递）
func getPreviousOutput(ctx *WorkflowContext) any {
	var last any
	for _, v := range ctx.NodeOutput {
		last = v
	}
	return last
}

func toMapAny(v any) map[string]any {
	if v == nil {
		return nil
	}
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return map[string]any{"result": v}
}

