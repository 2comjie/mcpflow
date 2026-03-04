package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"strings"
	"text/template"

	"github.com/2comjie/mcpflow/internal/llm"
	"github.com/2comjie/mcpflow/internal/mcp"
	"github.com/2comjie/mcpflow/internal/model"
	"github.com/dop251/goja"
	"github.com/expr-lang/expr"
	lua "github.com/yuin/gopher-lua"
)

// NodeExecutor 节点执行器接口
type NodeExecutor interface {
	Execute(ctx context.Context, node *model.Node, input map[string]any) (map[string]any, error)
}

// ExecutorRegistry 执行器注册表
type ExecutorRegistry struct {
	executors map[model.NodeType]NodeExecutor
}

func NewExecutorRegistry(mcpClient *mcp.Client) *ExecutorRegistry {
	r := &ExecutorRegistry{
		executors: make(map[model.NodeType]NodeExecutor),
	}
	r.executors[model.NodeStart] = &startExecutor{}
	r.executors[model.NodeEnd] = &endExecutor{}
	r.executors[model.NodeCondition] = &conditionExecutor{}
	r.executors[model.NodeCode] = &codeExecutor{}
	r.executors[model.NodeLLM] = &llmExecutor{}
	r.executors[model.NodeMCP] = &mcpExecutor{client: mcpClient}
	r.executors[model.NodeHTTP] = &httpExecutor{}
	r.executors[model.NodeEmail] = &emailExecutor{}
	return r
}

func (r *ExecutorRegistry) Get(nodeType model.NodeType) (NodeExecutor, error) {
	e, ok := r.executors[nodeType]
	if !ok {
		return nil, fmt.Errorf("no executor for node type: %s", nodeType)
	}
	return e, nil
}

// ==================== Start Executor ====================

type startExecutor struct{}

func (e *startExecutor) Execute(_ context.Context, _ *model.Node, input map[string]any) (map[string]any, error) {
	return input, nil
}

// ==================== End Executor ====================

type endExecutor struct{}

func (e *endExecutor) Execute(_ context.Context, _ *model.Node, input map[string]any) (map[string]any, error) {
	return input, nil
}

// ==================== Condition Executor ====================

type conditionExecutor struct{}

func (e *conditionExecutor) Execute(_ context.Context, node *model.Node, input map[string]any) (map[string]any, error) {
	if node.Config.Condition == nil {
		return nil, fmt.Errorf("condition config is nil")
	}

	program, err := expr.Compile(node.Config.Condition.Expression, expr.Env(input))
	if err != nil {
		return nil, fmt.Errorf("compile expression: %w", err)
	}

	result, err := expr.Run(program, input)
	if err != nil {
		return nil, fmt.Errorf("evaluate expression: %w", err)
	}

	branch := "false"
	switch v := result.(type) {
	case bool:
		if v {
			branch = "true"
		}
	default:
		branch = fmt.Sprintf("%v", v)
	}

	return map[string]any{"branch": branch}, nil
}

// ==================== Code Executor ====================

type codeExecutor struct{}

func (e *codeExecutor) Execute(_ context.Context, node *model.Node, input map[string]any) (map[string]any, error) {
	if node.Config.Code == nil {
		return nil, fmt.Errorf("code config is nil")
	}

	switch node.Config.Code.Language {
	case "javascript":
		return e.executeJS(node.Config.Code.Code, input)
	case "lua":
		return e.executeLua(node.Config.Code.Code, input)
	default:
		return nil, fmt.Errorf("unsupported language: %s", node.Config.Code.Language)
	}
}

func (e *codeExecutor) executeJS(code string, input map[string]any) (map[string]any, error) {
	vm := goja.New()
	if err := vm.Set("input", input); err != nil {
		return nil, fmt.Errorf("set input: %w", err)
	}

	wrapped := fmt.Sprintf("(function() { %s })()", code)
	val, err := vm.RunString(wrapped)
	if err != nil {
		return nil, fmt.Errorf("execute js: %w", err)
	}

	return convertJSResult(val)
}

func convertJSResult(val goja.Value) (map[string]any, error) {
	if val == nil || goja.IsUndefined(val) || goja.IsNull(val) {
		return map[string]any{}, nil
	}

	exported := val.Export()
	switch v := exported.(type) {
	case map[string]any:
		return v, nil
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return map[string]any{"result": v}, nil
		}
		var m map[string]any
		if err := json.Unmarshal(b, &m); err != nil {
			return map[string]any{"result": v}, nil
		}
		return m, nil
	}
}

func (e *codeExecutor) executeLua(code string, input map[string]any) (map[string]any, error) {
	L := lua.NewState()
	defer L.Close()

	inputTable := L.NewTable()
	for k, v := range input {
		setLuaValue(L, inputTable, k, v)
	}
	L.SetGlobal("input", inputTable)

	if err := L.DoString(code); err != nil {
		return nil, fmt.Errorf("execute lua: %w", err)
	}

	result := L.GetGlobal("result")
	return convertLuaResult(result), nil
}

func setLuaValue(L *lua.LState, table *lua.LTable, key string, value any) {
	switch v := value.(type) {
	case string:
		table.RawSetString(key, lua.LString(v))
	case float64:
		table.RawSetString(key, lua.LNumber(v))
	case bool:
		table.RawSetString(key, lua.LBool(v))
	case nil:
		table.RawSetString(key, lua.LNil)
	case map[string]any:
		sub := L.NewTable()
		for k, val := range v {
			setLuaValue(L, sub, k, val)
		}
		table.RawSetString(key, sub)
	}
}

func convertLuaResult(val lua.LValue) map[string]any {
	if val == nil || val == lua.LNil {
		return map[string]any{}
	}
	switch v := val.(type) {
	case *lua.LTable:
		result := make(map[string]any)
		v.ForEach(func(key, value lua.LValue) {
			if k, ok := key.(lua.LString); ok {
				result[string(k)] = luaValueToGo(value)
			}
		})
		return result
	default:
		return map[string]any{"result": luaValueToGo(val)}
	}
}

func luaValueToGo(val lua.LValue) any {
	switch v := val.(type) {
	case *lua.LNilType:
		return nil
	case lua.LBool:
		return bool(v)
	case lua.LNumber:
		return float64(v)
	case lua.LString:
		return string(v)
	case *lua.LTable:
		result := make(map[string]any)
		v.ForEach(func(key, value lua.LValue) {
			if k, ok := key.(lua.LString); ok {
				result[string(k)] = luaValueToGo(value)
			}
		})
		return result
	default:
		return v.String()
	}
}

// ==================== LLM Executor ====================

type llmExecutor struct{}

func (e *llmExecutor) Execute(ctx context.Context, node *model.Node, _ map[string]any) (map[string]any, error) {
	cfg := node.Config.LLM
	if cfg == nil {
		return nil, fmt.Errorf("llm config is nil")
	}

	client := llm.NewClient(cfg.BaseURL, cfg.APIKey)

	messages := make([]llm.Message, 0, 2)
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
		return nil, fmt.Errorf("llm chat: %w", err)
	}

	content := ""
	if len(resp.Choices) > 0 {
		content = resp.Choices[0].Message.Content
	}

	return map[string]any{
		"content":      content,
		"model":        resp.ID,
		"total_tokens": resp.Usage.TotalTokens,
	}, nil
}

// ==================== MCP Executor ====================

type mcpExecutor struct {
	client *mcp.Client
}

func (e *mcpExecutor) Execute(ctx context.Context, node *model.Node, _ map[string]any) (map[string]any, error) {
	cfg := node.Config.MCP
	if cfg == nil {
		return nil, fmt.Errorf("mcp config is nil")
	}

	switch cfg.Action {
	case "call_tool":
		return e.client.CallTool(ctx, cfg.ServerURL, cfg.ToolName, cfg.Arguments, cfg.Headers)
	case "get_prompt":
		return e.client.GetPrompt(ctx, cfg.ServerURL, cfg.PromptName, cfg.PromptArgs, cfg.Headers)
	case "read_resource":
		return e.client.ReadResource(ctx, cfg.ServerURL, cfg.ResourceURI, cfg.Headers)
	default:
		return nil, fmt.Errorf("unknown mcp action: %s", cfg.Action)
	}
}

// ==================== HTTP Executor ====================

type httpExecutor struct{}

func (e *httpExecutor) Execute(ctx context.Context, node *model.Node, _ map[string]any) (map[string]any, error) {
	cfg := node.Config.HTTP
	if cfg == nil {
		return nil, fmt.Errorf("http config is nil")
	}

	var bodyReader *bytes.Reader
	if cfg.Body != "" {
		bodyReader = bytes.NewReader([]byte(cfg.Body))
	} else {
		bodyReader = bytes.NewReader(nil)
	}

	req, err := http.NewRequestWithContext(ctx, cfg.Method, cfg.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	result := map[string]any{
		"status_code": resp.StatusCode,
		"body":        string(respBody),
	}

	var jsonBody map[string]any
	if err := json.Unmarshal(respBody, &jsonBody); err == nil {
		result["json"] = jsonBody
	}

	return result, nil
}

// ==================== Email Executor ====================

type emailExecutor struct{}

func (e *emailExecutor) Execute(_ context.Context, node *model.Node, input map[string]any) (map[string]any, error) {
	cfg := node.Config.Email
	if cfg == nil {
		return nil, fmt.Errorf("email config is nil")
	}

	// 渲染 Subject 和 Body 中的模板变量
	subject, err := renderTemplate(cfg.Subject, input)
	if err != nil {
		return nil, fmt.Errorf("render subject template: %w", err)
	}
	body, err := renderTemplate(cfg.Body, input)
	if err != nil {
		return nil, fmt.Errorf("render body template: %w", err)
	}

	contentType := cfg.ContentType
	if contentType == "" {
		contentType = "text/html"
	}

	toAddrs := splitAddrs(cfg.To)
	ccAddrs := splitAddrs(cfg.Cc)

	// 构建 MIME 邮件
	var msg bytes.Buffer
	msg.WriteString("From: " + cfg.From + "\r\n")
	msg.WriteString("To: " + cfg.To + "\r\n")
	if cfg.Cc != "" {
		msg.WriteString("Cc: " + cfg.Cc + "\r\n")
	}
	msg.WriteString("Subject: " + subject + "\r\n")
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: " + contentType + "; charset=\"UTF-8\"\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(body)

	// 收件人列表 = To + Cc
	allRecipients := append(toAddrs, ccAddrs...)

	addr := fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.SMTPHost)

	if err := smtp.SendMail(addr, auth, cfg.From, allRecipients, msg.Bytes()); err != nil {
		return nil, fmt.Errorf("send email: %w", err)
	}

	return map[string]any{
		"status":  "sent",
		"to":      cfg.To,
		"subject": subject,
	}, nil
}

func renderTemplate(tmplStr string, data map[string]any) (string, error) {
	if !strings.Contains(tmplStr, "{{") {
		return tmplStr, nil
	}
	tmpl, err := template.New("").Parse(tmplStr)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func splitAddrs(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	addrs := make([]string, 0, len(parts))
	for _, p := range parts {
		if a := strings.TrimSpace(p); a != "" {
			addrs = append(addrs, a)
		}
	}
	return addrs
}
