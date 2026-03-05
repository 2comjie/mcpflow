// 调试 Agent 执行流程：测试 LLM 是否会调用 MCP 工具
// 用法: LLM_BASE_URL=xxx LLM_API_KEY=xxx LLM_MODEL=xxx go run ./cmd/debug_agent
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	baseURL := os.Getenv("LLM_BASE_URL")
	apiKey := os.Getenv("LLM_API_KEY")
	model := os.Getenv("LLM_MODEL")
	if baseURL == "" || apiKey == "" || model == "" {
		log.Fatal("请设置环境变量: LLM_BASE_URL, LLM_API_KEY, LLM_MODEL")
	}

	// Step 1: 从 MCP 获取工具
	log.Println("=== Step 1: 获取 MCP 工具 ===")
	mcpTools := listMCPTools()
	log.Printf("发现 %d 个工具", len(mcpTools))

	// Step 2: 转换为 OpenAI function calling 格式
	var tools []map[string]any
	for i, t := range mcpTools {
		tm := t.(map[string]any)
		name := fmt.Sprintf("s0_%s", tm["name"])
		tools = append(tools, map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        name,
				"description": tm["description"],
				"parameters":  tm["inputSchema"],
			},
		})
		log.Printf("  Tool %d: %s - %s", i, name, tm["description"])
	}

	// Step 3: 构建消息并调用 LLM
	log.Println("\n=== Step 2: 调用 LLM (带 tools) ===")
	messages := []map[string]any{
		{"role": "system", "content": "你是一个智能助手，请使用可用的工具来完成用户的任务。"},
		{"role": "user", "content": "请查询北京的天气"},
	}

	reqBody := map[string]any{
		"model":       model,
		"messages":    messages,
		"tools":       tools,
		"temperature": 0.3,
		"max_tokens":  1024,
	}

	bodyJSON, _ := json.MarshalIndent(reqBody, "", "  ")
	log.Printf("请求 body:\n%s\n", string(bodyJSON))

	resp := callLLM(baseURL, apiKey, reqBody)
	respJSON, _ := json.MarshalIndent(resp, "", "  ")
	log.Printf("LLM 响应:\n%s\n", string(respJSON))

	// Step 4: 检查是否有 tool_calls
	choices, _ := resp["choices"].([]any)
	if len(choices) == 0 {
		log.Fatal("LLM 返回了 0 个 choices")
	}

	choice := choices[0].(map[string]any)
	msg := choice["message"].(map[string]any)
	finishReason, _ := choice["finish_reason"].(string)
	content, _ := msg["content"].(string)
	toolCalls, hasTC := msg["tool_calls"].([]any)

	log.Println("\n=== 结果分析 ===")
	log.Printf("finish_reason: %s", finishReason)
	log.Printf("content: %q", content)
	log.Printf("has tool_calls: %v", hasTC)
	if hasTC {
		log.Printf("tool_calls 数量: %d", len(toolCalls))
		for i, tc := range toolCalls {
			tcMap := tc.(map[string]any)
			fn := tcMap["function"].(map[string]any)
			log.Printf("  [%d] %s(%s)", i, fn["name"], fn["arguments"])
		}
		log.Println("\n✅ LLM 正确返回了 tool_calls，Agent 会调用 MCP 工具！")
	} else {
		log.Println("\n❌ LLM 没有返回 tool_calls，Agent 不会调用 MCP 工具")
		log.Println("可能原因:")
		log.Println("  1. 该 LLM 模型不支持 function calling")
		log.Println("  2. tools 格式不被该模型识别")
		log.Println("  3. 模型选择直接回答而不使用工具")
	}
}

func listMCPTools() []any {
	body := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/list",
	}
	bodyJSON, _ := json.Marshal(body)
	resp, err := http.Post("http://localhost:3002/mcp", "application/json", bytes.NewReader(bodyJSON))
	if err != nil {
		log.Fatalf("MCP 连接失败: %v (请先启动 go run ./cmd/testmcp)", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	var result map[string]any
	json.Unmarshal(data, &result)
	r := result["result"].(map[string]any)
	return r["tools"].([]any)
}

func callLLM(baseURL, apiKey string, reqBody map[string]any) map[string]any {
	bodyJSON, _ := json.Marshal(reqBody)
	url := baseURL + "/chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyJSON))
	if err != nil {
		log.Fatalf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("LLM 请求失败: %v", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		log.Fatalf("LLM 返回 %d: %s", resp.StatusCode, string(data))
	}

	var result map[string]any
	json.Unmarshal(data, &result)
	return result
}
