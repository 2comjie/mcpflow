package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
)

type Client struct {
	http  *http.Client
	idSeq atomic.Int64
}

func NewClient() *Client {
	return &Client{http: &http.Client{}}
}

// JSON-RPC 2.0 请求/响应

type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int64  `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *rpcError) Error() string {
	return fmt.Sprintf("rpc error %d: %s", e.Code, e.Message)
}

func (c *Client) call(ctx context.Context, serverURL, method string, params any, headers map[string]string) (json.RawMessage, error) {
	req := rpcRequest{
		JSONRPC: "2.0",
		ID:      c.idSeq.Add(1),
		Method:  method,
		Params:  params,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, serverURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(respBody))
	}

	var rpcResp rpcResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	if rpcResp.Error != nil {
		return nil, rpcResp.Error
	}
	return rpcResp.Result, nil
}

// CallTool 调用 MCP Server 的工具
func (c *Client) CallTool(ctx context.Context, serverURL, toolName string, args map[string]any, headers map[string]string) (map[string]any, error) {
	params := map[string]any{
		"name":      toolName,
		"arguments": args,
	}
	raw, err := c.call(ctx, serverURL, "tools/call", params, headers)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("unmarshal tool result: %w", err)
	}
	return result, nil
}

// GetPrompt 获取 MCP Server 的提示词
func (c *Client) GetPrompt(ctx context.Context, serverURL, promptName string, args map[string]any, headers map[string]string) (map[string]any, error) {
	params := map[string]any{
		"name":      promptName,
		"arguments": args,
	}
	raw, err := c.call(ctx, serverURL, "prompts/get", params, headers)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("unmarshal prompt result: %w", err)
	}
	return result, nil
}

// ReadResource 读取 MCP Server 的资源
func (c *Client) ReadResource(ctx context.Context, serverURL, uri string, headers map[string]string) (map[string]any, error) {
	params := map[string]any{
		"uri": uri,
	}
	raw, err := c.call(ctx, serverURL, "resources/read", params, headers)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("unmarshal resource result: %w", err)
	}
	return result, nil
}

// ListTools 列出 MCP Server 的所有工具
func (c *Client) ListTools(ctx context.Context, serverURL string, headers map[string]string) ([]any, error) {
	raw, err := c.call(ctx, serverURL, "tools/list", nil, headers)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Tools []any `json:"tools"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal tools list: %w", err)
	}
	return resp.Tools, nil
}

// ListPrompts 列出 MCP Server 的所有提示词
func (c *Client) ListPrompts(ctx context.Context, serverURL string, headers map[string]string) ([]any, error) {
	raw, err := c.call(ctx, serverURL, "prompts/list", nil, headers)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Prompts []any `json:"prompts"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal prompts list: %w", err)
	}
	return resp.Prompts, nil
}

// ListResources 列出 MCP Server 的所有资源
func (c *Client) ListResources(ctx context.Context, serverURL string, headers map[string]string) ([]any, error) {
	raw, err := c.call(ctx, serverURL, "resources/list", nil, headers)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Resources []any `json:"resources"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal resources list: %w", err)
	}
	return resp.Resources, nil
}

// TestConnection 测试与 MCP Server 的连接，返回能力信息
func (c *Client) TestConnection(ctx context.Context, serverURL string, headers map[string]string) (map[string]any, error) {
	raw, err := c.call(ctx, serverURL, "initialize", map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{},
		"clientInfo": map[string]any{
			"name":    "mcpflow",
			"version": "1.0.0",
		},
	}, headers)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("unmarshal init result: %w", err)
	}
	return result, nil
}
