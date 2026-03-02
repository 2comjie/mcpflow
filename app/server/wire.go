package main

import (
	"github.com/2comjie/mcpflow/internal/config"
	"github.com/2comjie/mcpflow/internal/mcp"
	"github.com/2comjie/mcpflow/internal/mcpserver"
	"github.com/2comjie/mcpflow/internal/secret"
	"github.com/2comjie/mcpflow/internal/workflow"
	"gorm.io/gorm"
)

// App 持有所有模块的 handler，供路由注册使用
type App struct {
	Workflow  *workflow.WorkflowHandler
	MCPServer *mcpserver.Handler
	Secret    *secret.Handler
}

func InitApp(cfg *config.Config, db *gorm.DB) *App {
	// MCP Client
	mcpClient := mcp.NewClient()

	// MCP Server 管理
	mcpSvc := mcpserver.NewService(db, mcpClient)
	mustMigrate(mcpSvc.AutoMigrate())

	// Secret
	secretRepo := secret.NewRepository(db)
	mustMigrate(secretRepo.AutoMigrate())

	// Workflow
	workflowRepo := workflow.NewWorkflowRepository(db)
	mustMigrate(workflowRepo.AutoMigrate())

	registry := workflow.NewExecutorRegistry()
	engine := workflow.NewEngine(registry)
	engine.SetSecretStore(secretRepo)
	workflowSvc := workflow.NewWorkflowService(workflowRepo, engine)

	return &App{
		Workflow:  workflow.NewWorkflowHandler(workflowSvc),
		MCPServer: mcpserver.NewHandler(mcpSvc),
		Secret:    secret.NewHandler(secretRepo),
	}
}

func mustMigrate(err error) {
	if err != nil {
		panic("failed to migrate " + err.Error())
	}
}
