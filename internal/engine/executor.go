package engine

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"strings"
	"text/template"
	"time"

	"log"

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
	r.executors[model.NodeAgent] = &agentExecutor{mcpClient: mcpClient}
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

// ==================== Agent Executor (LLM + MCP Tool Calling) ====================

type agentExecutor struct {
	mcpClient *mcp.Client
}

// NewAgentExecutor 创建 Agent 执行器（供外部直接调用，如 Agent Playground）
func NewAgentExecutor(mcpClient *mcp.Client) NodeExecutor {
	return &agentExecutor{mcpClient: mcpClient}
}

func (e *agentExecutor) Execute(ctx context.Context, node *model.Node, input map[string]any) (map[string]any, error) {
	cfg := node.Config.Agent
	if cfg == nil {
		return nil, fmt.Errorf("agent config is nil")
	}
	if len(cfg.McpServers) == 0 {
		return nil, fmt.Errorf("agent has no mcp servers configured")
	}

	maxIter := cfg.MaxIterations
	if maxIter <= 0 {
		maxIter = 10
	}

	// 步骤收集（用于可视化）
	var steps []map[string]any
	stepNum := 0

	// 1. 从所有 MCP Server 获取 tools 并转换为 OpenAI Tool 格式
	type serverTool struct {
		serverIdx int
		origName  string
	}
	toolMap := make(map[string]serverTool) // prefixed name -> server info
	var tools []llm.Tool

	for i, srv := range cfg.McpServers {
		log.Printf("[Agent] Discovering tools from MCP server: %s", srv.URL)
		t0 := time.Now()
		mcpTools, err := e.mcpClient.ListTools(ctx, srv.URL, srv.Headers)
		elapsed := time.Since(t0).Milliseconds()
		if err != nil {
			return nil, fmt.Errorf("list tools from %s: %w", srv.URL, err)
		}
		log.Printf("[Agent] Discovered %d tools from %s (%dms)", len(mcpTools), srv.URL, elapsed)

		var toolNames []string
		for _, t := range mcpTools {
			tm, ok := t.(map[string]any)
			if !ok {
				continue
			}
			name, _ := tm["name"].(string)
			desc, _ := tm["description"].(string)
			params := tm["inputSchema"]

			prefixed := fmt.Sprintf("s%d_%s", i, name)
			toolMap[prefixed] = serverTool{serverIdx: i, origName: name}
			toolNames = append(toolNames, name)

			tools = append(tools, llm.Tool{
				Type: "function",
				Function: llm.ToolFunction{
					Name:        prefixed,
					Description: desc,
					Parameters:  params,
				},
			})
		}

		stepNum++
		steps = append(steps, map[string]any{
			"step":       stepNum,
			"type":       "tool_discovery",
			"server_url": srv.URL,
			"tools":      toolNames,
			"duration":   elapsed,
		})
	}

	// 2. 构建初始 messages
	client := llm.NewClient(cfg.BaseURL, cfg.APIKey)
	messages := make([]llm.Message, 0, 4)
	if cfg.SystemMsg != "" {
		messages = append(messages, llm.Message{Role: "system", Content: cfg.SystemMsg})
	}
	messages = append(messages, llm.Message{Role: "user", Content: cfg.Prompt})

	// 3. 循环调用 LLM，处理 tool_calls
	totalToolCalls := 0
	iterations := 0
	var totalTokens int

	for iterations < maxIter {
		iterations++

		t0 := time.Now()
		resp, err := client.Chat(ctx, &llm.ChatRequest{
			Model:       cfg.Model,
			Messages:    messages,
			Tools:       tools,
			Temperature: cfg.Temperature,
			MaxTokens:   cfg.MaxTokens,
		})
		llmElapsed := time.Since(t0).Milliseconds()
		if err != nil {
			return nil, fmt.Errorf("agent llm chat (iter %d): %w", iterations, err)
		}

		if len(resp.Choices) == 0 {
			return nil, fmt.Errorf("agent llm returned no choices")
		}

		choice := resp.Choices[0]
		totalTokens += resp.Usage.TotalTokens
		log.Printf("[Agent] LLM iter=%d, tool_calls=%d, finish_reason=%s, content_len=%d",
			iterations, len(choice.Message.ToolCalls), choice.FinishReason, len(choice.Message.Content))

		// 记录 LLM 调用步骤
		stepNum++
		steps = append(steps, map[string]any{
			"step":             stepNum,
			"type":             "llm_call",
			"messages_count":   len(messages),
			"tools_count":      len(tools),
			"has_tool_calls":   len(choice.Message.ToolCalls) > 0,
			"tool_calls_count": len(choice.Message.ToolCalls),
			"content_preview":  truncateStr(choice.Message.Content, 200),
			"duration":         llmElapsed,
		})

		// 没有 tool_calls，返回最终结果
		if len(choice.Message.ToolCalls) == 0 {
			stepNum++
			steps = append(steps, map[string]any{
				"step":            stepNum,
				"type":            "final_response",
				"content_preview": truncateStr(choice.Message.Content, 500),
				"duration":        0,
			})
			return map[string]any{
				"content":          choice.Message.Content,
				"tool_calls_count": totalToolCalls,
				"iterations":       iterations,
				"total_tokens":     totalTokens,
				"agent_steps":      steps,
			}, nil
		}

		// 有 tool_calls，追加 assistant 消息
		log.Printf("[Agent] Processing %d tool calls", len(choice.Message.ToolCalls))
		messages = append(messages, choice.Message)

		// 执行每个 tool call
		for _, tc := range choice.Message.ToolCalls {
			totalToolCalls++
			st, ok := toolMap[tc.Function.Name]
			if !ok {
				messages = append(messages, llm.Message{
					Role:       "tool",
					ToolCallID: tc.ID,
					Content:    fmt.Sprintf("error: unknown tool %s", tc.Function.Name),
				})
				continue
			}

			// 解析参数
			var args map[string]any
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
				messages = append(messages, llm.Message{
					Role:       "tool",
					ToolCallID: tc.ID,
					Content:    fmt.Sprintf("error: invalid arguments: %v", err),
				})
				continue
			}

			// 通过 MCP 调用工具
			srv := cfg.McpServers[st.serverIdx]
			t0 := time.Now()
			result, err := e.mcpClient.CallTool(ctx, srv.URL, st.origName, args, srv.Headers)
			toolElapsed := time.Since(t0).Milliseconds()

			stepNum++
			toolStep := map[string]any{
				"step":      stepNum,
				"type":      "tool_call",
				"tool_name": st.origName,
				"tool_args": args,
				"duration":  toolElapsed,
			}

			if err != nil {
				toolStep["tool_error"] = err.Error()
				messages = append(messages, llm.Message{
					Role:       "tool",
					ToolCallID: tc.ID,
					Content:    fmt.Sprintf("error: %v", err),
				})
			} else {
				// 从 MCP 结果中提取纯文本（MCP 格式: {"content":[{"type":"text","text":"..."}]}）
				toolText := extractMCPText(result)
				toolStep["tool_result"] = truncateStr(toolText, 500)
				messages = append(messages, llm.Message{
					Role:       "tool",
					ToolCallID: tc.ID,
					Content:    toolText,
				})
			}
			steps = append(steps, toolStep)
		}
	}

	// 达到最大迭代次数，返回最后一次的内容
	return map[string]any{
		"content":          "Agent reached max iterations without final response",
		"tool_calls_count": totalToolCalls,
		"iterations":       iterations,
		"agent_steps":      steps,
	}, nil
}

// extractMCPText 从 MCP 工具调用结果中提取纯文本内容
// MCP 格式: {"content": [{"type":"text","text":"实际内容"}, ...]}
// 返回拼接后的纯文本，供 LLM 理解
func extractMCPText(result map[string]any) string {
	contentArr, ok := result["content"].([]any)
	if !ok {
		// 不是标准 MCP 格式，回退为 JSON
		b, _ := json.Marshal(result)
		return string(b)
	}
	var parts []string
	for _, item := range contentArr {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if text, ok := m["text"].(string); ok {
			parts = append(parts, text)
		}
	}
	if len(parts) == 0 {
		b, _ := json.Marshal(result)
		return string(b)
	}
	return strings.Join(parts, "\n")
}

// truncateStr 截断字符串，避免步骤数据过大
func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
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

	if err := sendMail(addr, auth, cfg.From, allRecipients, msg.Bytes(), cfg.SMTPHost, cfg.SMTPPort); err != nil {
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

// sendMail 发送邮件，支持 SSL (465) 和 STARTTLS (587/25)
func sendMail(addr string, auth smtp.Auth, from string, to []string, msg []byte, host string, port int) error {
	if port == 465 {
		// 隐式 TLS：QQ邮箱、163邮箱等使用 465 端口
		tlsConfig := &tls.Config{ServerName: host}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("tls dial: %w", err)
		}
		client, err := smtp.NewClient(conn, host)
		if err != nil {
			conn.Close()
			return fmt.Errorf("smtp client: %w", err)
		}
		defer client.Close()

		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("auth: %w", err)
		}
		if err := client.Mail(from); err != nil {
			return fmt.Errorf("mail from: %w", err)
		}
		for _, addr := range to {
			if err := client.Rcpt(addr); err != nil {
				return fmt.Errorf("rcpt to %s: %w", addr, err)
			}
		}
		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("data: %w", err)
		}
		if _, err := w.Write(msg); err != nil {
			return fmt.Errorf("write: %w", err)
		}
		if err := w.Close(); err != nil {
			return fmt.Errorf("close data: %w", err)
		}
		return client.Quit()
	}

	// STARTTLS：587/25 端口，使用标准库
	return smtp.SendMail(addr, auth, from, to, msg)
}
