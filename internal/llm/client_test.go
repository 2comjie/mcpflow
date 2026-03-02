package llm

import (
	"context"
	"os"
	"testing"
)

// 使用 DeepSeek API 做真实调用测试
// 运行: go test ./internal/llm/ -v -run TestDeepSeekChat
func TestDeepSeekChat(t *testing.T) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		t.Skip("DEEPSEEK_API_KEY not set, skipping integration test")
	}

	client := NewClient("https://api.deepseek.com", apiKey)

	resp, err := client.Chat(context.Background(), &ChatRequest{
		Model: "deepseek-chat",
		Messages: []Message{
			{Role: "system", Content: "你是一个helpful的助手，请简短回答。"},
			{Role: "user", Content: "1+1等于几？"},
		},
		Temperature: 0.7,
		MaxTokens:   100,
	})
	if err != nil {
		t.Fatalf("chat failed: %v", err)
	}

	if len(resp.Choices) == 0 {
		t.Fatal("no choices in response")
	}

	content := resp.Choices[0].Message.Content
	t.Logf("模型回复: %s", content)
	t.Logf("Token用量: prompt=%d, completion=%d, total=%d",
		resp.Usage.PromptTokens, resp.Usage.CompletionTokens, resp.Usage.TotalTokens)

	if content == "" {
		t.Error("response content is empty")
	}
}
