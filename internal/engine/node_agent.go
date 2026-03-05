package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/2comjie/mcpflow/internal/model"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type AgentResult struct {
	Content string           `json:"content"`
	Steps   []model.AgentStep `json:"steps"`
}

func executeAgent(cfg *model.AgentConfig, ctx *WorkflowContext) (any, error) {
	if cfg == nil {
		return nil, fmt.Errorf("agent config is nil")
	}

	maxIter := cfg.MaxIterations
	if maxIter <= 0 {
		maxIter = 10
	}

	// 收集所有 MCP 工具
	tools, mcpClients, err := collectMCPTools(cfg.McpServers)
	if err != nil {
		return nil, fmt.Errorf("collect mcp tools: %w", err)
	}
	defer func() {
		for _, c := range mcpClients {
			_ = c.Close()
		}
	}()

	// 构建工具名 -> MCP 客户端的映射
	toolClientMap := make(map[string]*client.Client)
	for name, c := range mcpClients {
		toolClientMap[name] = c
	}

	prompt := resolveTemplate(cfg.Prompt, ctx)
	systemMsg := resolveTemplate(cfg.SystemMsg, ctx)

	messages := []map[string]any{}
	if systemMsg != "" {
		messages = append(messages, map[string]any{"role": "system", "content": systemMsg})
	}
	messages = append(messages, map[string]any{"role": "user", "content": prompt})

	var steps []model.AgentStep

	for i := 0; i < maxIter; i++ {
		reqBody := map[string]any{
			"model":    cfg.Model,
			"messages": messages,
		}
		if len(tools) > 0 {
			reqBody["tools"] = tools
		}
		if cfg.Temperature > 0 {
			reqBody["temperature"] = cfg.Temperature
		}
		if cfg.MaxTokens > 0 {
			reqBody["max_tokens"] = cfg.MaxTokens
		}

		body, _ := json.Marshal(reqBody)
		url := cfg.BaseURL + "/chat/completions"
		req, err := http.NewRequest("POST", url, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("create agent request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("agent request: %w", err)
		}
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("agent api error (status %d): %s", resp.StatusCode, string(respBody))
		}

		var result map[string]any
		if err := json.Unmarshal(respBody, &result); err != nil {
			return nil, fmt.Errorf("parse agent response: %w", err)
		}

		choices, _ := result["choices"].([]any)
		if len(choices) == 0 {
			return nil, fmt.Errorf("empty choices in agent response")
		}
		choice, _ := choices[0].(map[string]any)
		message, _ := choice["message"].(map[string]any)
		finishReason, _ := choice["finish_reason"].(string)

		// 如果没有工具调用，返回最终内容
		toolCalls, hasToolCalls := message["tool_calls"].([]any)
		if !hasToolCalls || len(toolCalls) == 0 || finishReason == "stop" {
			content, _ := message["content"].(string)
			steps = append(steps, model.AgentStep{
				Iteration: i + 1,
				Type:      "response",
				Content:   content,
			})
			return AgentResult{Content: content, Steps: steps}, nil
		}

		// 将 assistant 消息加入历史
		messages = append(messages, message)

		// 处理工具调用
		for _, tc := range toolCalls {
			toolCall, _ := tc.(map[string]any)
			tcID, _ := toolCall["id"].(string)
			fn, _ := toolCall["function"].(map[string]any)
			toolName, _ := fn["name"].(string)
			argsStr, _ := fn["arguments"].(string)

			var args map[string]any
			_ = json.Unmarshal([]byte(argsStr), &args)

			steps = append(steps, model.AgentStep{
				Iteration: i + 1,
				Type:      "tool_call",
				ToolName:  toolName,
				ToolArgs:  args,
			})

			// 调用 MCP 工具
			toolResult, err := callMCPTool(toolClientMap, toolName, args)
			resultStr := ""
			if err != nil {
				resultStr = fmt.Sprintf("Error: %v", err)
			} else {
				resultStr = toolResult
			}

			steps[len(steps)-1].ToolResult = resultStr

			// 将工具结果加入消息历史
			messages = append(messages, map[string]any{
				"role":         "tool",
				"tool_call_id": tcID,
				"content":      resultStr,
			})
		}
	}

	return nil, fmt.Errorf("agent exceeded max iterations (%d)", maxIter)
}

// collectMCPTools 连接所有 MCP 服务器，收集工具列表
func collectMCPTools(servers []model.AgentMCPServer) ([]map[string]any, map[string]*client.Client, error) {
	var allTools []map[string]any
	clients := make(map[string]*client.Client)

	for _, srv := range servers {
		c, err := client.NewSSEMCPClient(srv.URL, client.WithHeaders(srv.Headers))
		if err != nil {
			return nil, nil, fmt.Errorf("connect mcp %s: %w", srv.URL, err)
		}

		if err := c.Start(context.Background()); err != nil {
			return nil, nil, fmt.Errorf("start mcp %s: %w", srv.URL, err)
		}

		initReq := mcp.InitializeRequest{}
		initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
		initReq.Params.ClientInfo = mcp.Implementation{Name: "mcpflow-agent", Version: "1.0"}
		if _, err := c.Initialize(context.Background(), initReq); err != nil {
			return nil, nil, fmt.Errorf("init mcp %s: %w", srv.URL, err)
		}

		toolsResult, err := c.ListTools(context.Background(), mcp.ListToolsRequest{})
		if err != nil {
			continue
		}

		for _, tool := range toolsResult.Tools {
			// 转为 OpenAI function calling 格式
			openaiTool := map[string]any{
				"type": "function",
				"function": map[string]any{
					"name":        tool.Name,
					"description": tool.Description,
					"parameters":  tool.InputSchema,
				},
			}
			allTools = append(allTools, openaiTool)
			clients[tool.Name] = c
		}
	}

	return allTools, clients, nil
}

// callMCPTool 调用 MCP 工具
func callMCPTool(clientMap map[string]*client.Client, toolName string, args map[string]any) (string, error) {
	c, ok := clientMap[toolName]
	if !ok {
		return "", fmt.Errorf("tool %q not found in any MCP server", toolName)
	}

	callReq := mcp.CallToolRequest{}
	callReq.Params.Name = toolName
	callReq.Params.Arguments = args

	result, err := c.CallTool(context.Background(), callReq)
	if err != nil {
		return "", err
	}

	// 提取文本内容
	var parts []string
	for _, content := range result.Content {
		if textContent, ok := content.(mcp.TextContent); ok {
			parts = append(parts, textContent.Text)
		}
	}

	if len(parts) > 0 {
		return joinStrings(parts), nil
	}

	// fallback：序列化整个结果
	b, _ := json.Marshal(result.Content)
	return string(b), nil
}

func joinStrings(ss []string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += "\n"
		}
		result += s
	}
	return result
}
