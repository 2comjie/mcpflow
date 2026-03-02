package workflow

import (
	"context"
	"fmt"

	"github.com/2comjie/mcpflow/internal/mcp"
)

// 执行器
type NodeExecutor interface {
	Execute(ctx context.Context, node *Node, input map[string]any) (map[string]any, error)
}

// 执行器注册表
type ExecutorRegistry struct {
	executors map[NodeType]NodeExecutor
}

func NewExecutorRegistry() *ExecutorRegistry {
	mcpClient := mcp.NewClient()

	r := &ExecutorRegistry{
		executors: make(map[NodeType]NodeExecutor),
	}
	r.Register(NodeStart, &StartExecutor{})
	r.Register(NodeEnd, &EndExecutor{})
	r.Register(NodeMCPTool, &MCPToolExecutor{client: mcpClient})
	r.Register(NodeMCPPrompt, &MCPPromptExecutor{client: mcpClient})
	r.Register(NodeMCPResource, &MCPResourceExecutor{client: mcpClient})
	r.Register(NodeLLM, &LLMExecutor{})
	r.Register(NodeCondition, &ConditionExecutor{})
	r.Register(NodeCode, &CodeExecutor{})
	r.Register(NodeHTTP, &HTTPExecutor{})
	return r
}

func (r *ExecutorRegistry) Register(nodeType NodeType, executor NodeExecutor) {
	r.executors[nodeType] = executor
}

func (r *ExecutorRegistry) Get(nodeType NodeType) (NodeExecutor, bool) {
	e, ok := r.executors[nodeType]
	if !ok {
		return nil, false
	}
	return e, true
}

// ==================== Start / End ====================

type StartExecutor struct{}

func (e *StartExecutor) Execute(ctx context.Context, node *Node, input map[string]any) (map[string]any, error) {
	return input, nil
}

type EndExecutor struct {
}

func (e *EndExecutor) Execute(ctx context.Context, node *Node, input map[string]any) (map[string]any, error) {
	return input, nil
}

// ==================== MCP 节点 ====================

type MCPToolExecutor struct {
	client *mcp.Client
}

func (e *MCPToolExecutor) Execute(ctx context.Context, node *Node, input map[string]any) (map[string]any, error) {
	cfg := node.Config.MCPTool
	if cfg == nil {
		return nil, fmt.Errorf("mcp_tool config is nil")
	}
	// 调用mcp server 的 tools/call
	return e.client.CallTool(ctx, cfg.ServerURL, cfg.ToolName, cfg.Arguments)
}

type MCPPromptExecutor struct {
	client *mcp.Client
}

func (e *MCPPromptExecutor) Execute(ctx context.Context, node *Node, input map[string]any) (map[string]any, error) {
	cfg := node.Config.MCPPrompt
	if cfg == nil {
		return nil, fmt.Errorf("mcp_prompt config is nil")
	}
	return e.client.GetPrompt(ctx, cfg.ServerURL, cfg.PromptName, cfg.Arguments)
}

type MCPResourceExecutor struct {
	client *mcp.Client
}

func (e *MCPResourceExecutor) Execute(ctx context.Context, node *Node, input map[string]any) (map[string]any, error) {
	// TODO: 从 node config 中获取 resource URI
	return nil, fmt.Errorf("mcp_resource not yet configured")
}

// ==================== LLM ====================

type LLMExecutor struct{}

func (e *LLMExecutor) Execute(ctx context.Context, node *Node, input map[string]any) (map[string]any, error) {
	cfg := node.Config.LLM
	if cfg == nil {
		return nil, fmt.Errorf("llm config is nil")
	}
	// TODO: 调用 LLM API (OpenAI / Anthropic 等)
	return map[string]any{
		"model":  cfg.Model,
		"result": "TODO: call llm api",
	}, nil
}

// ==================== Condition ====================

type ConditionExecutor struct{}

func (e *ConditionExecutor) Execute(ctx context.Context, node *Node, input map[string]any) (map[string]any, error) {
	cfg := node.Config.Condition
	if cfg == nil {
		return nil, fmt.Errorf("condition config is nil")
	}
	// TODO: 解析表达式，计算结果
	// 返回 branch 字段，引擎根据它选择下游
	return map[string]any{
		"branch": "true",
	}, nil
}

// ==================== Code ====================

type CodeExecutor struct{}

func (e *CodeExecutor) Execute(ctx context.Context, node *Node, input map[string]any) (map[string]any, error) {
	cfg := node.Config.Code
	if cfg == nil {
		return nil, fmt.Errorf("code config is nil")
	}
	// TODO: 沙箱执行代码
	return map[string]any{
		"result": "TODO: execute code",
	}, nil
}

// ==================== HTTP ====================

type HTTPExecutor struct{}

func (e *HTTPExecutor) Execute(ctx context.Context, node *Node, input map[string]any) (map[string]any, error) {
	cfg := node.Config.HTTP
	if cfg == nil {
		return nil, fmt.Errorf("http config is nil")
	}
	// TODO: 发起 HTTP 请求
	return map[string]any{
		"status": 200,
		"body":   "TODO: http call",
	}, nil
}
