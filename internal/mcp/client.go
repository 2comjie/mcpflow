package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

type Client struct {
	httpClient *http.Client
	idCounter  atomic.Int64
}

func NewClient() *Client {
	cl := &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	return cl
}

// JSON-RPC 2.0 请求/响应结构
type jsonRPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int64  `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (c *Client) call(ctx context.Context, serverURL string, method string, params any, headers map[string]string) (json.RawMessage, error) {
	reqBody := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      c.idCounter.Add(1),
		Method:  method,
		Params:  params,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, serverURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request %w", err)
	}
	defer resp.Body.Close()

	var rpcResp jsonRPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return nil, fmt.Errorf("decode response %w", err)
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("rpc error(%d) %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	return rpcResp.Result, nil
}

// Ping 简单健康检查，尝试调用 tools/list 验证连通性
func (c *Client) Ping(ctx context.Context, serverURL string, headers map[string]string) error {
	_, err := c.call(ctx, serverURL, "tools/list", nil, headers)
	return err
}

// 调用 MCP 工具 (tools/call)
func (c *Client) CallTool(ctx context.Context, serverURL, toolName string, arguments map[string]any, headers map[string]string) (map[string]any, error) {
	params := map[string]any{
		"name":      toolName,
		"arguments": arguments,
	}

	result, err := c.call(ctx, serverURL, "tools/call", params, headers)
	if err != nil {
		return nil, err
	}

	var out map[string]any
	if err := json.Unmarshal(result, &out); err != nil {
		return nil, fmt.Errorf("unmarshal tool result %w", err)
	}
	return out, nil
}

// 获取 MCP 提示词 (prompts/get)
func (c *Client) GetPrompt(ctx context.Context, serverURL, promptName string, arguments map[string]any, headers map[string]string) (map[string]any, error) {
	params := map[string]any{
		"name":      promptName,
		"arguments": arguments,
	}

	result, err := c.call(ctx, serverURL, "prompts/get", params, headers)
	if err != nil {
		return nil, err
	}

	var out map[string]any
	if err := json.Unmarshal(result, &out); err != nil {
		return nil, fmt.Errorf("unmarshal prompt result %w", err)
	}
	return out, nil
}

// 读取 MCP 资源 (resources/read)
func (c *Client) ReadResource(ctx context.Context, serverURL, uri string, headers map[string]string) (map[string]any, error) {
	params := map[string]any{
		"uri": uri,
	}

	result, err := c.call(ctx, serverURL, "resources/read", params, headers)
	if err != nil {
		return nil, err
	}

	var out map[string]any
	if err := json.Unmarshal(result, &out); err != nil {
		return nil, fmt.Errorf("unmarshal resource result %w", err)
	}
	return out, nil
}

// 列出 MCP 工具 (tools/list)
func (c *Client) ListTools(ctx context.Context, serverURL string, headers map[string]string) ([]map[string]any, error) {
	result, err := c.call(ctx, serverURL, "tools/list", nil, headers)
	if err != nil {
		return nil, err
	}

	var out struct {
		Tools []map[string]any `json:"tools"`
	}
	if err := json.Unmarshal(result, &out); err != nil {
		return nil, fmt.Errorf("unmarshal tools list: %w", err)
	}
	return out.Tools, nil
}

// 列出 MCP 提示词 (prompts/list)
func (c *Client) ListPrompts(ctx context.Context, serverURL string, headers map[string]string) ([]map[string]any, error) {
	result, err := c.call(ctx, serverURL, "prompts/list", nil, headers)
	if err != nil {
		return nil, err
	}

	var out struct {
		Prompts []map[string]any `json:"prompts"`
	}
	if err := json.Unmarshal(result, &out); err != nil {
		return nil, fmt.Errorf("unmarshal prompts list: %w", err)
	}
	return out.Prompts, nil
}

// 列出 MCP 资源 (resources/list)
func (c *Client) ListResources(ctx context.Context, serverURL string, headers map[string]string) ([]map[string]any, error) {
	result, err := c.call(ctx, serverURL, "resources/list", nil, headers)
	if err != nil {
		return nil, err
	}

	var out struct {
		Resources []map[string]any `json:"resources"`
	}
	if err := json.Unmarshal(result, &out); err != nil {
		return nil, fmt.Errorf("unmarshal resources list: %w", err)
	}
	return out.Resources, nil
}
