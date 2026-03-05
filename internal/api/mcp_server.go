package api

import (
	"context"
	"time"

	"github.com/2comjie/mcpflow/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/mark3labs/mcp-go/mcp"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (a *API) CreateMCPServer(c *gin.Context) {
	var srv model.MCPServer
	if err := c.ShouldBindJSON(&srv); err != nil {
		fail(c, 400, err.Error())
		return
	}
	if err := a.store.CreateMCPServer(&srv); err != nil {
		fail(c, 500, err.Error())
		return
	}
	ok(c, srv)
}

func (a *API) GetMCPServer(c *gin.Context) {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		fail(c, 400, "invalid id")
		return
	}
	srv, err := a.store.GetMCPServer(id)
	if err != nil {
		fail(c, 404, "mcp server not found")
		return
	}
	ok(c, srv)
}

func (a *API) ListMCPServers(c *gin.Context) {
	servers, err := a.store.ListMCPServers()
	if err != nil {
		fail(c, 500, err.Error())
		return
	}
	ok(c, servers)
}

func (a *API) UpdateMCPServer(c *gin.Context) {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		fail(c, 400, "invalid id")
		return
	}
	var updates map[string]any
	if err := c.ShouldBindJSON(&updates); err != nil {
		fail(c, 400, err.Error())
		return
	}
	delete(updates, "id")
	delete(updates, "_id")
	if err := a.store.UpdateMCPServer(id, updates); err != nil {
		fail(c, 500, err.Error())
		return
	}
	srv, _ := a.store.GetMCPServer(id)
	ok(c, srv)
}

func (a *API) DeleteMCPServer(c *gin.Context) {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		fail(c, 400, "invalid id")
		return
	}
	if err := a.store.DeleteMCPServer(id); err != nil {
		fail(c, 500, err.Error())
		return
	}
	ok(c, nil)
}

// CheckMCPServer 连接 MCP 服务器，获取工具/提示/资源列表并缓存
func (a *API) CheckMCPServer(c *gin.Context) {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		fail(c, 400, "invalid id")
		return
	}
	srv, err := a.store.GetMCPServer(id)
	if err != nil {
		fail(c, 404, "mcp server not found")
		return
	}

	mcpClient, err := connectMCPServer(srv.URL, srv.Headers)
	if err != nil {
		updateServerStatus(a, id, "error")
		fail(c, 500, "connect failed: "+err.Error())
		return
	}
	defer mcpClient.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 获取工具列表
	var tools any
	if toolsResult, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{}); err == nil {
		tools = toolsResult.Tools
	}

	// 获取提示列表
	var prompts any
	if promptsResult, err := mcpClient.ListPrompts(ctx, mcp.ListPromptsRequest{}); err == nil {
		prompts = promptsResult.Prompts
	}

	// 获取资源列表
	var resources any
	if resourcesResult, err := mcpClient.ListResources(ctx, mcp.ListResourcesRequest{}); err == nil {
		resources = resourcesResult.Resources
	}

	now := time.Now()
	updates := map[string]any{
		"status":     "active",
		"tools":      tools,
		"prompts":    prompts,
		"resources":  resources,
		"checked_at": now,
	}
	_ = a.store.UpdateMCPServer(id, updates)

	srv, _ = a.store.GetMCPServer(id)
	ok(c, srv)
}

func updateServerStatus(a *API, id bson.ObjectID, status string) {
	_ = a.store.UpdateMCPServer(id, map[string]any{"status": status})
}

// CallMCPTool 调用 MCP 服务器上的工具
func (a *API) CallMCPTool(c *gin.Context) {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		fail(c, 400, "invalid id")
		return
	}
	srv, err := a.store.GetMCPServer(id)
	if err != nil {
		fail(c, 404, "mcp server not found")
		return
	}

	var req struct {
		ToolName  string         `json:"tool_name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, err.Error())
		return
	}
	if req.ToolName == "" {
		fail(c, 400, "tool_name is required")
		return
	}

	mcpClient, err := connectMCPServer(srv.URL, srv.Headers)
	if err != nil {
		fail(c, 500, "connect failed: "+err.Error())
		return
	}
	defer mcpClient.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	callReq := mcp.CallToolRequest{}
	callReq.Params.Name = req.ToolName
	callReq.Params.Arguments = req.Arguments

	start := time.Now()
	result, err := mcpClient.CallTool(ctx, callReq)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		fail(c, 500, "call tool failed: "+err.Error())
		return
	}

	ok(c, gin.H{
		"result":   result.Content,
		"duration": duration,
	})
}
