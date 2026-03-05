package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

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

	var bodyReader io.Reader
	if cfg.Body != "" {
		bodyReader = strings.NewReader(cfg.Body)
	}

	req, err := http.NewRequest(method, cfg.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	result := map[string]any{
		"status":      resp.StatusCode,
		"status_text": resp.Status,
	}

	// 尝试解析 JSON
	var jsonBody any
	if err := json.Unmarshal(body, &jsonBody); err == nil {
		result["body"] = jsonBody
	} else {
		result["body"] = string(body)
	}

	return result, nil
}
