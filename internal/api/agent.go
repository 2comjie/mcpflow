package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/2comjie/mcpflow/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type AgentChatRequest struct {
	LLMProviderID string   `json:"llm_provider_id"`
	MCPServerIDs  []string `json:"mcp_server_ids"`
	Message       string   `json:"message"`
	SystemMsg     string   `json:"system_msg"`
	Model         string   `json:"model"`
	MaxIterations int      `json:"max_iterations"`
	Temperature   float64  `json:"temperature"`
	MaxTokens     int      `json:"max_tokens"`
}

func (a *API) AgentChat(c *gin.Context) {
	var req AgentChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, err.Error())
		return
	}
	if req.Message == "" {
		fail(c, 400, "message is required")
		return
	}

	// 获取 LLM Provider
	providerID, err := bson.ObjectIDFromHex(req.LLMProviderID)
	if err != nil {
		fail(c, 400, "invalid llm_provider_id")
		return
	}
	provider, err := a.store.GetLLMProvider(providerID)
	if err != nil {
		fail(c, 404, "llm provider not found")
		return
	}
	if provider.BaseURL == "" {
		fail(c, 400, "llm provider base_url is empty")
		return
	}

	// 确定模型
	modelName := req.Model
	if modelName == "" && len(provider.Models) > 0 {
		modelName = provider.Models[0]
	}
	if modelName == "" {
		fail(c, 400, "no model specified and provider has no models configured")
		return
	}

	// 获取 MCP Servers（可选）
	var mcpServers []model.AgentMCPServer
	for _, idStr := range req.MCPServerIDs {
		srvID, err := bson.ObjectIDFromHex(idStr)
		if err != nil {
			continue
		}
		srv, err := a.store.GetMCPServer(srvID)
		if err != nil {
			continue
		}
		mcpServers = append(mcpServers, model.AgentMCPServer{
			URL:     srv.URL,
			Headers: srv.Headers,
		})
	}

	maxIter := req.MaxIterations
	if maxIter <= 0 {
		maxIter = 10
	}

	// 收集 MCP 工具
	tools, mcpClients, err := collectAgentMCPTools(mcpServers)
	if err != nil {
		fail(c, 500, "collect mcp tools: "+err.Error())
		return
	}
	defer func() {
		for _, cl := range mcpClients {
			_ = cl.Close()
		}
	}()

	toolClientMap := make(map[string]*client.Client)
	for name, cl := range mcpClients {
		toolClientMap[name] = cl
	}

	// 构建消息
	messages := []map[string]any{}
	if req.SystemMsg != "" {
		messages = append(messages, map[string]any{"role": "system", "content": req.SystemMsg})
	}
	messages = append(messages, map[string]any{"role": "user", "content": req.Message})

	var steps []map[string]any
	totalTokens := 0
	toolCallsCount := 0

	for i := 0; i < maxIter; i++ {
		reqBody := map[string]any{
			"model":    modelName,
			"messages": messages,
		}
		if len(tools) > 0 {
			reqBody["tools"] = tools
		}
		if req.Temperature > 0 {
			reqBody["temperature"] = req.Temperature
		}
		if req.MaxTokens > 0 {
			reqBody["max_tokens"] = req.MaxTokens
		}

		body, _ := json.Marshal(reqBody)
		url := provider.BaseURL + "/chat/completions"
		httpReq, _ := http.NewRequest("POST", url, bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+provider.APIKey)

		steps = append(steps, map[string]any{
			"step":           len(steps) + 1,
			"type":           "llm_call",
			"messages_count": len(messages),
			"tools_count":    len(tools),
		})

		httpClient := &http.Client{Timeout: 120 * time.Second}
		resp, err := httpClient.Do(httpReq)
		if err != nil {
			fail(c, 500, "agent request: "+err.Error())
			return
		}
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fail(c, 500, fmt.Sprintf("agent api error (status %d): %s", resp.StatusCode, string(respBody)))
			return
		}

		var result map[string]any
		json.Unmarshal(respBody, &result)

		// 统计 token
		if usage, ok := result["usage"].(map[string]any); ok {
			if t, ok := usage["total_tokens"].(float64); ok {
				totalTokens += int(t)
			}
		}

		choices, _ := result["choices"].([]any)
		if len(choices) == 0 {
			fail(c, 500, "empty choices")
			return
		}
		choice, _ := choices[0].(map[string]any)
		message, _ := choice["message"].(map[string]any)
		finishReason, _ := choice["finish_reason"].(string)

		toolCalls, hasToolCalls := message["tool_calls"].([]any)
		if !hasToolCalls || len(toolCalls) == 0 || finishReason == "stop" {
			content, _ := message["content"].(string)
			contentPreview := content
			if len(contentPreview) > 200 {
				contentPreview = contentPreview[:200] + "..."
			}
			steps[len(steps)-1]["has_tool_calls"] = false
			steps[len(steps)-1]["content_preview"] = contentPreview

			steps = append(steps, map[string]any{
				"step":            len(steps) + 1,
				"type":            "final_response",
				"content_preview": contentPreview,
			})

			ok(c, gin.H{
				"content":          content,
				"agent_steps":      steps,
				"tool_calls_count": toolCallsCount,
				"iterations":       i + 1,
				"total_tokens":     totalTokens,
			})
			return
		}

		steps[len(steps)-1]["has_tool_calls"] = true
		steps[len(steps)-1]["tool_calls_count"] = len(toolCalls)

		messages = append(messages, message)

		for _, tc := range toolCalls {
			toolCall, _ := tc.(map[string]any)
			tcID, _ := toolCall["id"].(string)
			fn, _ := toolCall["function"].(map[string]any)
			toolName, _ := fn["name"].(string)
			argsStr, _ := fn["arguments"].(string)

			var args map[string]any
			json.Unmarshal([]byte(argsStr), &args)

			toolCallsCount++
			stepIdx := len(steps) + 1

			start := time.Now()
			toolResult, callErr := callAgentMCPTool(toolClientMap, toolName, args)
			duration := time.Since(start).Milliseconds()

			resultStr := ""
			if callErr != nil {
				resultStr = fmt.Sprintf("Error: %v", callErr)
			} else {
				resultStr = toolResult
			}

			steps = append(steps, map[string]any{
				"step":        stepIdx,
				"type":        "tool_call",
				"tool_name":   toolName,
				"tool_args":   args,
				"tool_result":  resultStr,
				"tool_error":  "",
				"duration":    duration,
			})
			if callErr != nil {
				steps[len(steps)-1]["tool_error"] = callErr.Error()
			}

			messages = append(messages, map[string]any{
				"role":         "tool",
				"tool_call_id": tcID,
				"content":      resultStr,
			})
		}
	}

	fail(c, 500, fmt.Sprintf("agent exceeded max iterations (%d)", maxIter))
}

// connectMCPServer connects to an MCP server, trying Streamable HTTP first, then falling back to SSE.
func connectMCPServer(url string, headers map[string]string) (*client.Client, error) {
	ctx := context.Background()

	// Try Streamable HTTP first (modern protocol)
	var httpOpts []transport.StreamableHTTPCOption
	if len(headers) > 0 {
		httpOpts = append(httpOpts, transport.WithHTTPHeaders(headers))
	}
	c, err := client.NewStreamableHttpClient(url, httpOpts...)
	if err == nil {
		if startErr := c.Start(ctx); startErr == nil {
			initReq := mcp.InitializeRequest{}
			initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
			initReq.Params.ClientInfo = mcp.Implementation{Name: "mcpflow-agent", Version: "1.0"}
			if _, initErr := c.Initialize(ctx, initReq); initErr == nil {
				slog.Info("connected to MCP server via Streamable HTTP", "url", url)
				return c, nil
			}
			_ = c.Close()
		} else {
			_ = c.Close()
		}
	}

	// Fallback to SSE transport
	sseURL := url
	if !strings.HasSuffix(sseURL, "/sse") {
		sseURL = strings.TrimRight(sseURL, "/") + "/sse"
	}
	c, err = client.NewSSEMCPClient(sseURL, client.WithHeaders(headers))
	if err != nil {
		return nil, fmt.Errorf("connect mcp %s: %w", url, err)
	}
	if err := c.Start(ctx); err != nil {
		_ = c.Close()
		return nil, fmt.Errorf("start mcp %s: %w", url, err)
	}
	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{Name: "mcpflow-agent", Version: "1.0"}
	if _, err := c.Initialize(ctx, initReq); err != nil {
		_ = c.Close()
		return nil, fmt.Errorf("init mcp %s: %w", url, err)
	}
	slog.Info("connected to MCP server via SSE", "url", sseURL)
	return c, nil
}

func collectAgentMCPTools(servers []model.AgentMCPServer) ([]map[string]any, map[string]*client.Client, error) {
	var allTools []map[string]any
	clients := make(map[string]*client.Client)

	cleanup := func() {
		for _, c := range clients {
			_ = c.Close()
		}
	}

	for _, srv := range servers {
		c, err := connectMCPServer(srv.URL, srv.Headers)
		if err != nil {
			cleanup()
			return nil, nil, err
		}
		toolsResult, err := c.ListTools(context.Background(), mcp.ListToolsRequest{})
		if err != nil {
			continue
		}
		for _, tool := range toolsResult.Tools {
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

func callAgentMCPTool(clientMap map[string]*client.Client, toolName string, args map[string]any) (string, error) {
	c, ok := clientMap[toolName]
	if !ok {
		return "", fmt.Errorf("tool %q not found", toolName)
	}
	callReq := mcp.CallToolRequest{}
	callReq.Params.Name = toolName
	callReq.Params.Arguments = args
	result, err := c.CallTool(context.Background(), callReq)
	if err != nil {
		return "", err
	}
	var parts []string
	for _, content := range result.Content {
		if textContent, ok := content.(mcp.TextContent); ok {
			parts = append(parts, textContent.Text)
		}
	}
	if len(parts) > 0 {
		out := ""
		for i, s := range parts {
			if i > 0 {
				out += "\n"
			}
			out += s
		}
		return out, nil
	}
	b, _ := json.Marshal(result.Content)
	return string(b), nil
}
