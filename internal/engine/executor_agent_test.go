package engine

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/2comjie/mcpflow/internal/mcp"
	"github.com/2comjie/mcpflow/internal/model"
)

// TestAgentFullFlow 完整测试 Agent 执行流程：
// 1. Agent 从 MCP 发现工具
// 2. Agent 调用 LLM，LLM 返回 tool_calls
// 3. Agent 通过 MCP 调用工具
// 4. Agent 将工具结果传回 LLM
// 5. LLM 返回最终文本响应
func TestAgentFullFlow(t *testing.T) {
	// --- Mock MCP Server ---
	mcpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Method string          `json:"method"`
			ID     any             `json:"id"`
			Params json.RawMessage `json:"params"`
		}
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &req)

		t.Logf("[MCP] method=%s", req.Method)

		var result any
		switch req.Method {
		case "tools/list":
			result = map[string]any{
				"tools": []map[string]any{
					{
						"name":        "get_weather",
						"description": "获取指定城市的天气信息",
						"inputSchema": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"city": map[string]any{
									"type":        "string",
									"description": "城市名称",
								},
							},
							"required": []string{"city"},
						},
					},
				},
			}
		case "tools/call":
			var p struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments"`
			}
			json.Unmarshal(req.Params, &p)
			t.Logf("[MCP] tools/call: name=%s args=%v", p.Name, p.Arguments)
			result = map[string]any{
				"content": []map[string]any{
					{"type": "text", "text": `{"city":"Beijing","temperature":38,"weather":"晴"}`},
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"jsonrpc": "2.0",
			"id":      req.ID,
			"result":  result,
		})
	}))
	defer mcpServer.Close()

	// --- Mock LLM Server ---
	llmCallCount := 0
	llmServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		json.Unmarshal(body, &req)

		llmCallCount++
		messages := req["messages"].([]any)
		tools := req["tools"]

		t.Logf("[LLM] Call #%d: %d messages, tools=%v", llmCallCount, len(messages), tools != nil)

		// 打印发送给 LLM 的请求体
		prettyBody, _ := json.MarshalIndent(req, "", "  ")
		t.Logf("[LLM] Request body:\n%s", string(prettyBody))

		var resp map[string]any
		if llmCallCount == 1 {
			// 第一次调用：LLM 返回 tool_calls
			resp = map[string]any{
				"id": "test-1",
				"choices": []map[string]any{
					{
						"index": 0,
						"message": map[string]any{
							"role":    "assistant",
							"content": "",
							"tool_calls": []map[string]any{
								{
									"id":   "call_1",
									"type": "function",
									"function": map[string]any{
										"name":      "s0_get_weather",
										"arguments": `{"city":"Beijing"}`,
									},
								},
							},
						},
						"finish_reason": "tool_calls",
					},
				},
				"usage": map[string]any{
					"prompt_tokens":     100,
					"completion_tokens": 20,
					"total_tokens":      120,
				},
			}
		} else {
			// 第二次调用：LLM 返回最终文本（已获得工具结果）
			resp = map[string]any{
				"id": "test-2",
				"choices": []map[string]any{
					{
						"index": 0,
						"message": map[string]any{
							"role":    "assistant",
							"content": "北京今日天气晴朗，温度38°C。",
						},
						"finish_reason": "stop",
					},
				},
				"usage": map[string]any{
					"prompt_tokens":     200,
					"completion_tokens": 30,
					"total_tokens":      230,
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer llmServer.Close()

	// --- 运行 Agent ---
	mcpClient := mcp.NewClient()
	executor := &agentExecutor{mcpClient: mcpClient}

	node := &model.Node{
		ID:   "test_agent",
		Type: model.NodeAgent,
		Name: "Test Agent",
		Config: model.NodeConfig{
			Agent: &model.AgentConfig{
				BaseURL:       llmServer.URL,
				APIKey:        "test-key",
				Model:         "test-model",
				Prompt:        "请查询北京的天气",
				SystemMsg:     "你是一个智能助手，请使用可用的工具来完成用户的任务。",
				MaxIterations: 10,
				Temperature:   0.3,
				MaxTokens:     1024,
				McpServers: []model.AgentMCPServer{
					{URL: mcpServer.URL},
				},
			},
		},
	}

	result, err := executor.Execute(context.Background(), node, map[string]any{})
	if err != nil {
		t.Fatalf("Agent execution failed: %v", err)
	}

	// --- 验证结果 ---
	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	t.Logf("Agent result:\n%s", string(resultJSON))

	// 验证 LLM 被调用了 2 次（一次返回 tool_calls，一次返回最终结果）
	if llmCallCount != 2 {
		t.Errorf("Expected 2 LLM calls, got %d", llmCallCount)
	}

	// 验证最终内容
	content, ok := result["content"].(string)
	if !ok || content == "" {
		t.Error("Expected non-empty content in result")
	}
	t.Logf("Final content: %s", content)

	// 验证 tool_calls 计数
	toolCallsCount, _ := result["tool_calls_count"].(int)
	if toolCallsCount != 1 {
		t.Errorf("Expected 1 tool call, got %d", toolCallsCount)
	}

	// 验证 agent_steps
	steps, ok := result["agent_steps"].([]map[string]any)
	if !ok {
		t.Fatal("Expected agent_steps in result")
	}
	t.Logf("Total steps: %d", len(steps))

	// 应该有 4 步: tool_discovery, llm_call(with tool_calls), tool_call, llm_call(final), final_response
	expectedTypes := []string{"tool_discovery", "llm_call", "tool_call", "llm_call", "final_response"}
	if len(steps) != len(expectedTypes) {
		t.Errorf("Expected %d steps, got %d", len(expectedTypes), len(steps))
		for i, s := range steps {
			t.Logf("  step %d: type=%s", i, s["type"])
		}
	} else {
		for i, s := range steps {
			if s["type"] != expectedTypes[i] {
				t.Errorf("Step %d: expected type %s, got %s", i, expectedTypes[i], s["type"])
			}
		}
	}
}
