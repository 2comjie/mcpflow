package main

import (
	"fmt"
	"log"
	"os"

	"github.com/2comjie/mcpflow/internal/api"
	"github.com/2comjie/mcpflow/internal/config"
	"github.com/2comjie/mcpflow/internal/engine"
	"github.com/2comjie/mcpflow/internal/mcp"
	"github.com/2comjie/mcpflow/internal/store"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	cfgPath := "configs/config.yml"
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		cfgPath = p
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := gorm.Open(mysql.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}

	s := store.New(db)
	if err := s.AutoMigrate(); err != nil {
		log.Fatalf("auto migrate: %v", err)
	}

	mcpClient := mcp.NewClient()
	registry := engine.NewExecutorRegistry(mcpClient)
	eng := engine.NewEngine(registry)

	app := api.New(s, eng, mcpClient)

	r := gin.Default()
	app.RegisterRoutes(r.Group("/api/v1"))

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
