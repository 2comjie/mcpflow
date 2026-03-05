package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/2comjie/mcpflow/internal/model"
)

func executeLLM(cfg *model.LLMConfig, ctx *WorkflowContext) (any, error) {
	if cfg == nil {
		return nil, fmt.Errorf("llm config is nil")
	}

	prompt := resolveTemplate(cfg.Prompt, ctx)
	systemMsg := resolveTemplate(cfg.SystemMsg, ctx)

	messages := []map[string]string{}
	if systemMsg != "" {
		messages = append(messages, map[string]string{"role": "system", "content": systemMsg})
	}
	messages = append(messages, map[string]string{"role": "user", "content": prompt})

	reqBody := map[string]any{
		"model":    cfg.Model,
		"messages": messages,
	}
	if cfg.Temperature > 0 {
		reqBody["temperature"] = cfg.Temperature
	}
	if cfg.MaxTokens > 0 {
		reqBody["max_tokens"] = cfg.MaxTokens
	}

	// 结构化输出：通过 response_format 指定 JSON Schema
	if cfg.OutputSchema != nil {
		reqBody["response_format"] = map[string]any{
			"type": "json_schema",
			"json_schema": map[string]any{
				"name":   "structured_output",
				"strict": true,
				"schema": cfg.OutputSchema,
			},
		}
	}

	body, _ := json.Marshal(reqBody)
	url := cfg.BaseURL + "/chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create llm request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	httpClient := &http.Client{Timeout: 120 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("llm request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("llm api error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse llm response: %w", err)
	}

	content := extractContent(result)

	// 如果有 output_schema，解析为结构化 JSON
	if cfg.OutputSchema != nil {
		var structured any
		if err := json.Unmarshal([]byte(content), &structured); err != nil {
			return nil, fmt.Errorf("parse structured output: %w", err)
		}
		return structured, nil
	}

	return map[string]any{"content": content}, nil
}

func extractContent(result map[string]any) string {
	choices, _ := result["choices"].([]any)
	if len(choices) == 0 {
		return ""
	}
	choice, _ := choices[0].(map[string]any)
	msg, _ := choice["message"].(map[string]any)
	content, _ := msg["content"].(string)
	return content
}
