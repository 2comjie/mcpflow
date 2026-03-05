package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/2comjie/mcpflow/internal/model"
)

func executeHTTP(cfg *model.HTTPConfig, ctx *WorkflowContext) (any, error) {
	if cfg == nil {
		return nil, fmt.Errorf("http config is nil")
	}

	method := strings.ToUpper(cfg.Method)
	if method == "" {
		method = "GET"
	}

	url := resolveTemplate(cfg.URL, ctx)
	body := resolveTemplate(cfg.Body, ctx)

	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}

	httpClient := &http.Client{Timeout: 60 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	result := map[string]any{
		"status":      resp.StatusCode,
		"status_text": resp.Status,
	}

	// 尝试解析 JSON
	var jsonBody any
	if err := json.Unmarshal(respBody, &jsonBody); err == nil {
		result["body"] = jsonBody
	} else {
		result["body"] = string(respBody)
	}

	return result, nil
}
