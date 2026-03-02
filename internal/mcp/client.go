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

func (c *Client) call(ctx context.Context, serverURL string, method string, params any) (json.RawMessage, error) {
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

// 调用 MCP 工具 (tools/call)
func (c *Client) CallTool(ctx context.Context, serverURL, toolName string, arguments map[string]any) (map[string]any, error) {
	params := map[string]any{
		"name":      toolName,
		"arguments": arguments,
	}

	result, err := c.call(ctx, serverURL, "tools/call", params)
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
func (c *Client) GetPrompt(ctx context.Context, serverURL, promptName string, arguments map[string]any) (map[string]any, error) {
	params := map[string]any{
		"name":      promptName,
		"arguments": arguments,
	}

	result, err := c.call(ctx, serverURL, "prompts/get", params)
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
func (c *Client) ReadResource(ctx context.Context, serverURL, uri string) (map[string]any, error) {
	params := map[string]any{
		"uri": uri,
	}

	result, err := c.call(ctx, serverURL, "resources/read", params)
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
func (c *Client) ListTools(ctx context.Context, serverURL string) ([]map[string]any, error) {
	result, err := c.call(ctx, serverURL, "tools/list", nil)
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
