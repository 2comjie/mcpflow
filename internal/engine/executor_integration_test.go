// 集成测试：测试真实 LLM 和邮件节点
// 运行: go test -v -run TestIntegration -tags=integration ./internal/engine/
//
//go:build integration

package engine

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/2comjie/mcpflow/internal/model"
)

// TestIntegrationLLMNode 测试 LLM 节点调用 DeepSeek API
func TestIntegrationLLMNode(t *testing.T) {
	executor := &llmExecutor{}

	node := &model.Node{
		ID:   "test_llm",
		Type: model.NodeLLM,
		Name: "DeepSeek LLM",
		Config: model.NodeConfig{
			LLM: &model.LLMConfig{
				BaseURL:     "https://api.deepseek.com/v1",
				APIKey:      "sk-853775c33f6e4570817830bc86e62151",
				Model:       "deepseek-chat",
				Prompt:      "用一句话介绍什么是工作流引擎",
				SystemMsg:   "你是一个技术专家，请简洁回答",
				Temperature: 0.3,
				MaxTokens:   256,
			},
		},
	}

	result, err := executor.Execute(context.Background(), node, map[string]any{})
	if err != nil {
		t.Fatalf("LLM execution failed: %v", err)
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	t.Logf("LLM result:\n%s", string(resultJSON))

	content, ok := result["content"].(string)
	if !ok || content == "" {
		t.Error("Expected non-empty content from LLM")
	}
	t.Logf("LLM response: %s", content)
}

// TestIntegrationEmailNode 测试邮件节点通过 QQ 邮箱 SMTP 发送
func TestIntegrationEmailNode(t *testing.T) {
	executor := &emailExecutor{}

	node := &model.Node{
		ID:   "test_email",
		Type: model.NodeEmail,
		Name: "QQ Email",
		Config: model.NodeConfig{
			Email: &model.EmailConfig{
				SMTPHost:    "smtp.qq.com",
				SMTPPort:    465,
				Username:    "2comjie@qq.com",
				Password:    "wbzcxscbfwgbdhac",
				From:        "2comjie@qq.com",
				To:          "2comjie@qq.com",
				Subject:     "MCPFlow 集成测试 - 邮件节点",
				Body:        "<h3>MCPFlow 邮件节点测试</h3><p>这是一封来自 MCPFlow 工作流引擎集成测试的邮件。</p><p>如果你收到这封邮件，说明邮件节点功能正常。</p>",
				ContentType: "text/html",
			},
		},
	}

	result, err := executor.Execute(context.Background(), node, map[string]any{})
	if err != nil {
		t.Fatalf("Email execution failed: %v", err)
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	t.Logf("Email result:\n%s", string(resultJSON))

	status, _ := result["status"].(string)
	if status != "sent" {
		t.Errorf("Expected status 'sent', got '%s'", status)
	}
	t.Log("Email sent successfully!")
}

// TestIntegrationLLMThenEmail 测试工作流：LLM 生成内容 → 邮件发送
func TestIntegrationLLMThenEmail(t *testing.T) {
	// Step 1: LLM 生成内容
	llmExec := &llmExecutor{}
	llmNode := &model.Node{
		ID:   "llm_node",
		Type: model.NodeLLM,
		Name: "生成邮件内容",
		Config: model.NodeConfig{
			LLM: &model.LLMConfig{
				BaseURL:     "https://api.deepseek.com/v1",
				APIKey:      "sk-853775c33f6e4570817830bc86e62151",
				Model:       "deepseek-chat",
				Prompt:      "请写一段50字以内的每日技术分享，主题是Go语言的并发模型，格式用HTML",
				SystemMsg:   "你是一个技术博主",
				Temperature: 0.7,
				MaxTokens:   512,
			},
		},
	}

	llmResult, err := llmExec.Execute(context.Background(), llmNode, map[string]any{})
	if err != nil {
		t.Fatalf("LLM execution failed: %v", err)
	}

	content, _ := llmResult["content"].(string)
	t.Logf("LLM generated content:\n%s", content)

	if content == "" {
		t.Fatal("LLM returned empty content")
	}

	// Step 2: 将 LLM 生成的内容通过邮件发送
	emailExec := &emailExecutor{}
	emailNode := &model.Node{
		ID:   "email_node",
		Type: model.NodeEmail,
		Name: "发送邮件",
		Config: model.NodeConfig{
			Email: &model.EmailConfig{
				SMTPHost:    "smtp.qq.com",
				SMTPPort:    465,
				Username:    "2comjie@qq.com",
				Password:    "wbzcxscbfwgbdhac",
				From:        "2comjie@qq.com",
				To:          "2comjie@qq.com",
				Subject:     "MCPFlow 每日技术分享 - LLM 自动生成",
				Body:        content,
				ContentType: "text/html",
			},
		},
	}

	emailResult, err := emailExec.Execute(context.Background(), emailNode, map[string]any{})
	if err != nil {
		t.Fatalf("Email execution failed: %v", err)
	}

	t.Logf("Email result: %v", emailResult)
	t.Log("LLM → Email workflow test passed!")
}
