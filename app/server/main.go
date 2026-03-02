package main

import (
	"fmt"
	"log"

	"github.com/2comjie/mcpflow/internal/config"
	"github.com/2comjie/mcpflow/internal/llm"
	"github.com/2comjie/mcpflow/internal/mcp"
	"github.com/2comjie/mcpflow/internal/mcpserver"
	"github.com/2comjie/mcpflow/internal/workflow"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 加载配置
	cfg, err := config.Load("configs/config.yml")
	if err != nil {
		log.Fatalf("failed to load config %v", err)
	}

	// 连接 MySQL
	db, err := gorm.Open(mysql.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database %v", err)
	}

	// 初始化 workflow 模块
	repo := workflow.NewWorkflowRepository(db)
	if err := repo.AutoMigrate(); err != nil {
		log.Fatalf("failed to migrate %v", err)
	}

	mcpClient := mcp.NewClient()
	mcpSvc := mcpserver.NewService(db, mcpClient)
	if err := mcpSvc.AutoMigrate(); err != nil {
		log.Fatalf("failed to migrate %v", err)
	}

	// 初始化 LLM Client
	llmClient := llm.NewClient(cfg.LLM.BaseURL, cfg.LLM.APIKey)

	// 传入 registry
	registry := workflow.NewExecutorRegistry(llmClient)
	engine := workflow.NewEngine(registry)
	svc := workflow.NewWorkflowService(repo, engine)
	handler := workflow.NewWorkflowHandler(svc)

	// 启动 Gin
	r := gin.Default()
	api := r.Group("/api/v1")
	handler.RegisterRoutes(api)

	mcpHandler := mcpserver.NewHandler(mcpSvc)
	mcpHandler.RegisterRoutes(api)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
