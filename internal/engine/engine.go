package engine

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
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

	// 构建反向边：target -> []source 信息
	reverseEdges := buildReverseEdgeMap(wf.Edges)
	edgeMap := buildEdgeMap(wf.Edges)

	// ctx 存储节点间传递的数据
	ctx := &WorkflowContext{
		Input:      input,
		NodeOutput: make(map[string]any),
		EdgeMap:    edgeMap,
		ExecOrder:  order,
	}

	skipped := make(map[string]bool)

	for _, nodeID := range order {
		node := nodeMap[nodeID]
		if node == nil {
			continue
		}

		// 检查是否应该跳过（条件分支控制）
		if shouldSkip(nodeID, reverseEdges, ctx, skipped) {
			skipped[nodeID] = true
			exec.NodeStates[node.ID] = NodeState{
				NodeID: node.ID,
				Status: "skipped",
			}
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
		if agentRes, ok := output.(AgentResult); ok {
			log.AgentSteps = agentRes.Steps
			log.Output = map[string]any{"content": agentRes.Content}
			// Store as map so downstream nodes can access {{nodes.node_id.content}}
			output = map[string]any{
				"content":      agentRes.Content,
				"agent_steps":  agentRes.Steps,
			}
		}
		_ = e.store.CreateExecutionLog(log)

		if err != nil {
			return e.failExecution(exec, err)
		}

		ctx.NodeOutput[node.ID] = output
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

// NodeEvent 节点执行事件（用于 SSE 推送）
type NodeEvent struct {
	NodeID   string `json:"node_id"`
	NodeName string `json:"node_name"`
	NodeType string `json:"node_type"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
	Duration int64  `json:"duration,omitempty"`
}

// ExecuteWorkflowWithEvents 执行工作流并通过 channel 推送事件
func (e *Engine) ExecuteWorkflowWithEvents(wf *model.Workflow, input map[string]any, events chan<- NodeEvent) (*model.Execution, error) {
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

	reverseEdges := buildReverseEdgeMap(wf.Edges)
	edgeMap := buildEdgeMap(wf.Edges)

	ctx := &WorkflowContext{
		Input:      input,
		NodeOutput: make(map[string]any),
		EdgeMap:    edgeMap,
		ExecOrder:  order,
	}

	skipped := make(map[string]bool)

	for _, nodeID := range order {
		node := nodeMap[nodeID]
		if node == nil {
			continue
		}

		if shouldSkip(nodeID, reverseEdges, ctx, skipped) {
			skipped[nodeID] = true
			exec.NodeStates[node.ID] = NodeState{NodeID: node.ID, Status: "skipped"}
			continue
		}

		// 发送 running 事件
		if events != nil {
			events <- NodeEvent{
				NodeID:   node.ID,
				NodeName: node.Name,
				NodeType: string(node.Type),
				Status:   "running",
			}
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

		// 发送完成/失败事件
		if events != nil {
			evt := NodeEvent{
				NodeID:   node.ID,
				NodeName: node.Name,
				NodeType: string(node.Type),
				Status:   state.Status,
				Duration: duration,
			}
			if err != nil {
				evt.Error = err.Error()
			}
			events <- evt
		}

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
		if agentRes, ok := output.(AgentResult); ok {
			log.AgentSteps = agentRes.Steps
			log.Output = map[string]any{"content": agentRes.Content}
			output = map[string]any{
				"content":      agentRes.Content,
				"agent_steps":  agentRes.Steps,
			}
		}
		_ = e.store.CreateExecutionLog(log)

		if err != nil {
			return e.failExecution(exec, err)
		}
		ctx.NodeOutput[node.ID] = output
	}

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
	Input      map[string]any
	NodeOutput map[string]any
	EdgeMap    map[string][]EdgeInfo
	ExecOrder  []string // 拓扑排序后的节点顺序
}

type EdgeInfo struct {
	TargetID  string
	Condition string
}

type ReverseEdgeInfo struct {
	SourceID  string
	Condition string
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

func buildReverseEdgeMap(edges []model.Edge) map[string][]ReverseEdgeInfo {
	m := make(map[string][]ReverseEdgeInfo)
	for _, e := range edges {
		m[e.Target] = append(m[e.Target], ReverseEdgeInfo{
			SourceID:  e.Source,
			Condition: e.Condition,
		})
	}
	return m
}

// shouldSkip 判断节点是否应该被跳过
// 规则：如果入边带有条件标签（"true"/"false"），检查条件节点的输出是否匹配
func shouldSkip(nodeID string, reverseEdges map[string][]ReverseEdgeInfo, ctx *WorkflowContext, skipped map[string]bool) bool {
	inEdges := reverseEdges[nodeID]
	if len(inEdges) == 0 {
		return false
	}

	for _, edge := range inEdges {
		// 如果上游节点被跳过，当前节点也跳过
		if skipped[edge.SourceID] {
			return true
		}

		// 如果边有条件标签，检查条件节点的输出
		if edge.Condition != "" {
			output, exists := ctx.NodeOutput[edge.SourceID]
			if !exists {
				return true
			}
			condResult, ok := output.(bool)
			if !ok {
				continue
			}
			// edge.Condition 为 "true" 或 "false"
			if edge.Condition == "true" && !condResult {
				return true
			}
			if edge.Condition == "false" && condResult {
				return true
			}
		}
	}
	return false
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

// getPreviousOutput 获取上一个已执行节点的输出（按拓扑顺序）
func getPreviousOutput(ctx *WorkflowContext) any {
	for i := len(ctx.ExecOrder) - 1; i >= 0; i-- {
		nodeID := ctx.ExecOrder[i]
		if output, ok := ctx.NodeOutput[nodeID]; ok {
			return output
		}
	}
	return ctx.Input
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

// resolveTemplate 模板替换，支持 {{input.xxx}} 和 {{nodes.node_id.xxx}} 和 {{nodes.node_id}}
var tmplPattern = regexp.MustCompile(`\{\{(\w+(?:\.\w+)*)\}\}`)

func resolveTemplate(tmpl string, ctx *WorkflowContext) string {
	if tmpl == "" || !strings.Contains(tmpl, "{{") {
		return tmpl
	}

	return tmplPattern.ReplaceAllStringFunc(tmpl, func(match string) string {
		path := match[2 : len(match)-2] // 去掉 {{ 和 }}
		parts := strings.SplitN(path, ".", 2)

		switch parts[0] {
		case "input":
			if len(parts) == 1 {
				return jsonString(ctx.Input)
			}
			val := getNestedValue(ctx.Input, parts[1])
			if val != nil {
				return fmt.Sprintf("%v", val)
			}
		case "nodes":
			if len(parts) == 1 {
				return jsonString(ctx.NodeOutput)
			}
			// parts[1] 可以是 "node_id" 或 "node_id.field"
			subParts := strings.SplitN(parts[1], ".", 2)
			nodeOutput, ok := ctx.NodeOutput[subParts[0]]
			if !ok {
				return match
			}
			if len(subParts) == 1 {
				return jsonString(nodeOutput)
			}
			if m, ok := nodeOutput.(map[string]any); ok {
				val := getNestedValue(m, subParts[1])
				if val != nil {
					return fmt.Sprintf("%v", val)
				}
			}
		}
		return match
	})
}

func getNestedValue(m map[string]any, path string) any {
	parts := strings.Split(path, ".")
	var current any = m
	for _, p := range parts {
		cm, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current = cm[p]
	}
	return current
}

func jsonString(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}
