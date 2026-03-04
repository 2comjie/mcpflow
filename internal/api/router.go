package api

import (
	"github.com/2comjie/mcpflow/internal/engine"
	"github.com/2comjie/mcpflow/internal/mcp"
	"github.com/2comjie/mcpflow/internal/store"
	"github.com/gin-gonic/gin"
)

type API struct {
	store  *store.Store
	engine *engine.Engine
	mcp    *mcp.Client
}

func New(store *store.Store, engine *engine.Engine, mcp *mcp.Client) *API {
	return &API{store: store, engine: engine, mcp: mcp}
}

func (a *API) RegisterRoutes(r *gin.RouterGroup) {
	// Workflow
	wf := r.Group("/workflows")
	{
		wf.POST("", a.CreateWorkflow)
		wf.GET("", a.ListWorkflows)
		wf.GET("/:id", a.GetWorkflow)
		wf.PUT("/:id", a.UpdateWorkflow)
		wf.DELETE("/:id", a.DeleteWorkflow)
		wf.POST("/:id/execute", a.ExecuteWorkflow)
		wf.GET("/:id/executions", a.ListWorkflowExecutions)
	}

	// Execution
	exec := r.Group("/executions")
	{
		exec.GET("", a.ListExecutions)
		exec.GET("/:id", a.GetExecution)
		exec.GET("/:id/logs", a.GetExecutionLogs)
		exec.DELETE("/:id", a.DeleteExecution)
	}

	// MCP Server
	ms := r.Group("/mcp-servers")
	{
		ms.POST("", a.CreateMCPServer)
		ms.GET("", a.ListMCPServers)
		ms.PUT("/:id", a.UpdateMCPServer)
		ms.DELETE("/:id", a.DeleteMCPServer)
		ms.POST("/:id/test", a.TestMCPServer)
		ms.GET("/:id/tools", a.GetMCPServerTools)
		ms.GET("/:id/prompts", a.GetMCPServerPrompts)
		ms.GET("/:id/resources", a.GetMCPServerResources)
	}

	// LLM Provider
	lp := r.Group("/llm-providers")
	{
		lp.POST("", a.CreateLLMProvider)
		lp.GET("", a.ListLLMProviders)
		lp.PUT("/:id", a.UpdateLLMProvider)
		lp.DELETE("/:id", a.DeleteLLMProvider)
	}
}
