package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/2comjie/mcpflow/internal/api"
	"github.com/2comjie/mcpflow/internal/config"
	"github.com/2comjie/mcpflow/internal/engine"
	"github.com/2comjie/mcpflow/internal/mcp"
	"github.com/2comjie/mcpflow/internal/store"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
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

	client, err := mongo.Connect(options.Client().ApplyURI(cfg.Database.URI))
	if err != nil {
		log.Fatalf("connect mongodb: %v", err)
	}
	defer client.Disconnect(context.TODO())

	db := client.Database(cfg.Database.Name)
	s := store.New(db)

	mcpClient := mcp.NewClient()
	registry := engine.NewExecutorRegistry(mcpClient)
	eng := engine.NewEngine(registry)

	app := api.New(s, eng, mcpClient)

	r := gin.Default()
	app.RegisterRoutes(r.Group("/api/v1"))

	// 静态文件服务（生产模式：前端构建产物在 web/dist）
	webDist := "web/dist"
	if _, err := os.Stat(webDist); err == nil {
		r.Static("/assets", filepath.Join(webDist, "assets"))
		r.StaticFile("/favicon.ico", filepath.Join(webDist, "favicon.ico"))
		r.NoRoute(func(c *gin.Context) {
			// API 路径返回 404 JSON
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			// 其余路径返回 index.html（SPA 路由）
			c.File(filepath.Join(webDist, "index.html"))
		})
	}

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
