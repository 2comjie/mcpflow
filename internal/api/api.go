package api

import (
	"github.com/2comjie/mcpflow/internal/engine"
	"github.com/2comjie/mcpflow/internal/store"
	"github.com/gin-gonic/gin"
)

type API struct {
	store  *store.Store
	engine *engine.Engine
}

func New(s *store.Store, e *engine.Engine) *API {
	return &API{store: s, engine: e}
}

func (a *API) RegisterRoutes(r *gin.Engine) {
	v1 := r.Group("/api/v1")

	// 工作流
	wf := v1.Group("/workflows")
	wf.POST("", a.CreateWorkflow)
	wf.GET("", a.ListWorkflows)
	wf.GET("/:id", a.GetWorkflow)
	wf.PUT("/:id", a.UpdateWorkflow)
	wf.DELETE("/:id", a.DeleteWorkflow)
	wf.POST("/:id/execute", a.ExecuteWorkflow)

	// 执行记录
	exec := v1.Group("/executions")
	exec.GET("", a.ListExecutions)
	exec.GET("/:id", a.GetExecution)
	exec.GET("/:id/logs", a.GetExecutionLogs)
	exec.DELETE("/:id", a.DeleteExecution)

	// 工作流的执行记录
	wf.GET("/:id/executions", a.ListWorkflowExecutions)

	// MCP 服务器
	mcp := v1.Group("/mcp-servers")
	mcp.POST("", a.CreateMCPServer)
	mcp.GET("", a.ListMCPServers)
	mcp.GET("/:id", a.GetMCPServer)
	mcp.PUT("/:id", a.UpdateMCPServer)
	mcp.DELETE("/:id", a.DeleteMCPServer)
	mcp.POST("/:id/check", a.CheckMCPServer)

	// LLM Provider
	llm := v1.Group("/llm-providers")
	llm.POST("", a.CreateLLMProvider)
	llm.GET("", a.ListLLMProviders)
	llm.GET("/:id", a.GetLLMProvider)
	llm.PUT("/:id", a.UpdateLLMProvider)
	llm.DELETE("/:id", a.DeleteLLMProvider)

	// 统计
	v1.GET("/stats", a.GetStats)
}

// JSON 响应辅助
func ok(c *gin.Context, data any) {
	c.JSON(200, gin.H{"data": data})
}

func okList(c *gin.Context, data any, total int64) {
	c.JSON(200, gin.H{"data": data, "total": total})
}

func fail(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{"error": msg})
}
