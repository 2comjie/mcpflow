package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/2comjie/mcpflow/internal/model"
	"github.com/gin-gonic/gin"
)

func (a *API) CreateMCPServer(c *gin.Context) {
	var srv model.MCPServer
	if err := c.ShouldBindJSON(&srv); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := a.store.CreateMCPServer(&srv); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, srv)
}

func (a *API) ListMCPServers(c *gin.Context) {
	list, err := a.store.ListMCPServers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (a *API) UpdateMCPServer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var updates map[string]any
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := a.store.UpdateMCPServer(uint(id), updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (a *API) DeleteMCPServer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := a.store.DeleteMCPServer(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (a *API) TestMCPServer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	srv, err := a.store.GetMCPServer(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "server not found"})
		return
	}

	headers := srv.GetHeadersMap()

	// 测试连接
	if _, err := a.mcp.TestConnection(c.Request.Context(), srv.URL, headers); err != nil {
		now := time.Now()
		a.store.UpdateMCPServerCache(srv.ID, map[string]any{"status": "inactive", "checked_at": now})
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	// 获取能力列表并缓存
	tools, _ := a.mcp.ListTools(c.Request.Context(), srv.URL, headers)
	prompts, _ := a.mcp.ListPrompts(c.Request.Context(), srv.URL, headers)
	resources, _ := a.mcp.ListResources(c.Request.Context(), srv.URL, headers)

	now := time.Now()
	updates := map[string]any{
		"status":     "active",
		"tools":      tools,
		"prompts":    prompts,
		"resources":  resources,
		"checked_at": now,
	}
	a.store.UpdateMCPServerCache(srv.ID, updates)

	c.JSON(http.StatusOK, gin.H{
		"message":   "ok",
		"tools":     tools,
		"prompts":   prompts,
		"resources": resources,
	})
}

func (a *API) GetMCPServerTools(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	srv, err := a.store.GetMCPServer(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "server not found"})
		return
	}

	// 优先返回缓存
	if srv.Tools != nil {
		c.JSON(http.StatusOK, srv.Tools)
		return
	}

	headers := srv.GetHeadersMap()
	tools, err := a.mcp.ListTools(c.Request.Context(), srv.URL, headers)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	a.store.UpdateMCPServerCache(srv.ID, map[string]any{"tools": tools})
	c.JSON(http.StatusOK, tools)
}

func (a *API) GetMCPServerPrompts(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	srv, err := a.store.GetMCPServer(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "server not found"})
		return
	}

	if srv.Prompts != nil {
		c.JSON(http.StatusOK, srv.Prompts)
		return
	}

	headers := srv.GetHeadersMap()
	prompts, err := a.mcp.ListPrompts(c.Request.Context(), srv.URL, headers)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	a.store.UpdateMCPServerCache(srv.ID, map[string]any{"prompts": prompts})
	c.JSON(http.StatusOK, prompts)
}

func (a *API) GetMCPServerResources(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	srv, err := a.store.GetMCPServer(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "server not found"})
		return
	}

	if srv.Resources != nil {
		c.JSON(http.StatusOK, srv.Resources)
		return
	}

	headers := srv.GetHeadersMap()
	resources, err := a.mcp.ListResources(c.Request.Context(), srv.URL, headers)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	a.store.UpdateMCPServerCache(srv.ID, map[string]any{"resources": resources})
	c.JSON(http.StatusOK, resources)
}
