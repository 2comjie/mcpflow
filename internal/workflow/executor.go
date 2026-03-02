package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/2comjie/mcpflow/internal/llm"
	"github.com/2comjie/mcpflow/internal/mcp"
	"github.com/2comjie/mcpflow/pkg/httpx"
	"github.com/dop251/goja"
	"github.com/expr-lang/expr"
	lua "github.com/yuin/gopher-lua"
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
	r.Register(NodeHTTP, &HTTPExecutor{
		client: &http.Client{Timeout: 30 * time.Second},
	})
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
	cfg := node.Config.MCPResource
	if cfg == nil {
		return nil, fmt.Errorf("mcp_resource config is nil")
	}
	return e.client.ReadResource(ctx, cfg.ServerURL, cfg.URI)
}

// ==================== LLM ====================

type LLMExecutor struct{}

func (e *LLMExecutor) Execute(ctx context.Context, node *Node, input map[string]any) (map[string]any, error) {
	cfg := node.Config.LLM
	if cfg == nil {
		return nil, fmt.Errorf("llm config is nil")
	}
	if cfg.BaseURL == "" || cfg.APIKey == "" {
		return nil, fmt.Errorf("llm base_url and api_key are required")
	}

	client := llm.NewClient(cfg.BaseURL, cfg.APIKey)

	var messages []llm.Message
	if cfg.SystemMsg != "" {
		messages = append(messages, llm.Message{Role: "system", Content: cfg.SystemMsg})
	}
	messages = append(messages, llm.Message{Role: "user", Content: cfg.Prompt})

	resp, err := client.Chat(ctx, &llm.ChatRequest{
		Model:       cfg.Model,
		Messages:    messages,
		Temperature: cfg.Temperature,
		MaxTokens:   cfg.MaxTokens,
	})
	if err != nil {
		return nil, err
	}

	content := ""
	if len(resp.Choices) > 0 {
		content = resp.Choices[0].Message.Content
	}

	return map[string]any{
		"content":      content,
		"model":        cfg.Model,
		"total_tokens": resp.Usage.TotalTokens,
	}, nil
}

// ==================== Condition ====================

type ConditionExecutor struct{}

func (e *ConditionExecutor) Execute(ctx context.Context, node *Node, input map[string]any) (map[string]any, error) {
	cfg := node.Config.Condition
	if cfg == nil {
		return nil, fmt.Errorf("condition config is nil")
	}

	output, err := expr.Eval(cfg.Expression, input)
	if err != nil {
		return nil, fmt.Errorf("eval expression %w", err)
	}

	branch := "false"
	if b, ok := output.(bool); ok && b {
		branch = "true"
	}

	return map[string]any{"branch": branch}, nil
}

// ==================== Code ====================

type CodeExecutor struct{}

func (e *CodeExecutor) Execute(ctx context.Context, node *Node, input map[string]any) (map[string]any, error) {
	cfg := node.Config.Code
	if cfg == nil {
		return nil, fmt.Errorf("code config is nil")
	}

	switch cfg.Language {
	case "javascript", "js", "":
		return e.runJS(cfg.Code, input)
	case "lua":
		return e.runLua(cfg.Code, input)
	default:
		return nil, fmt.Errorf("unsupported language %s", cfg.Language)
	}
}

func (e *CodeExecutor) runJS(code string, input map[string]any) (map[string]any, error) {
	vm := goja.New()
	if err := vm.Set("input", input); err != nil {
		return nil, fmt.Errorf("set input %w", err)
	}

	val, err := vm.RunString(code)
	if err != nil {
		return nil, fmt.Errorf("execute js %w", err)
	}

	exported := val.Export()
	switch v := exported.(type) {
	case map[string]any:
		return v, nil
	default:
		return map[string]any{"result": exported}, nil
	}
}

func (e *CodeExecutor) runLua(code string, input map[string]any) (map[string]any, error) {
	L := lua.NewState()
	defer L.Close()

	// 注入 input 表
	L.SetGlobal("input", e.luaValueFromMap(L, input))

	if err := L.DoString(code); err != nil {
		return nil, fmt.Errorf("execute lua %w", err)
	}

	// 从全局变量 output 获取返回值
	result := L.GetGlobal("output")
	if result == lua.LNil {
		return map[string]any{"result": nil}, nil
	}

	if tbl, ok := result.(*lua.LTable); ok {
		return luaTableToMap(tbl), nil
	}
	return map[string]any{"result": luaToGoValue(result)}, nil
}

// 将 map[string]any 转为 LTable
func (e *CodeExecutor) luaValueFromMap(L *lua.LState, m map[string]any) *lua.LTable {
	tbl := L.NewTable()
	for k, v := range m {
		tbl.RawSetString(k, e.goToLuaValue(L, v))
	}
	return tbl
}

func (e *CodeExecutor) goToLuaValue(L *lua.LState, v any) lua.LValue {
	switch val := v.(type) {
	case string:
		return lua.LString(val)
	case int:
		return lua.LNumber(float64(val))
	case int64:
		return lua.LNumber(float64(val))
	case float64:
		return lua.LNumber(val)
	case bool:
		return lua.LBool(val)
	case map[string]any:
		return e.luaValueFromMap(L, val)
	case nil:
		return lua.LNil
	default:
		return lua.LString(fmt.Sprintf("%v", val))
	}
}

// luaTableToMap 将 LTable 转为 map[string]any
func luaTableToMap(tbl *lua.LTable) map[string]any {
	result := make(map[string]any)
	tbl.ForEach(func(k, v lua.LValue) {
		if key, ok := k.(lua.LString); ok {
			result[string(key)] = luaToGoValue(v)
		}
	})
	return result
}

func luaToGoValue(v lua.LValue) any {
	switch val := v.(type) {
	case lua.LBool:
		return bool(val)
	case lua.LNumber:
		return float64(val)
	case *lua.LNilType:
		return nil
	case lua.LString:
		return string(val)
	case *lua.LTable:
		return luaTableToMap(val)
	default:
		return val.String()
	}
}

// ==================== HTTP ====================

type HTTPExecutor struct {
	client *http.Client
}

func (e *HTTPExecutor) Execute(ctx context.Context, node *Node, input map[string]any) (map[string]any, error) {
	cfg := node.Config.HTTP
	if cfg == nil {
		return nil, fmt.Errorf("http config is nil")
	}

	var bodyReader io.Reader
	if cfg.Body != "" {
		bodyReader = strings.NewReader(cfg.Body)
	}

	req, err := http.NewRequestWithContext(ctx, cfg.Method, cfg.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request %w", err)
	}

	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response %w", err)
	}

	result := map[string]any{
		"status":  resp.StatusCode,
		"headers": httpx.HeaderToMap(resp.Header),
		"body":    string(respBody),
	}

	// 尝试解析 JSON 响应
	var jsonBody any
	if json.Unmarshal(respBody, &jsonBody) == nil {
		result["json"] = jsonBody
	}

	return result, nil
}
