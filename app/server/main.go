package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/2comjie/mcpflow/internal/api"
	"github.com/2comjie/mcpflow/internal/config"
	"github.com/2comjie/mcpflow/internal/engine"
	"github.com/2comjie/mcpflow/internal/store"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// slogMiddleware 用 slog 记录请求信息
func slogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		slog.Info("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration", duration.String(),
			"ip", c.ClientIP(),
		)
	}
}

func main() {
	// 初始化 slog
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	cfgPath := "configs/config.yml"
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		cfgPath = p
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		slog.Error("load config failed", "error", err)
		os.Exit(1)
	}

	client, err := mongo.Connect(options.Client().ApplyURI(cfg.Database.URI))
	if err != nil {
		slog.Error("connect mongodb failed", "error", err)
		os.Exit(1)
	}
	defer client.Disconnect(context.TODO())

	db := client.Database(cfg.Database.Name)
	s := store.New(db)
	e := engine.New(s)
	a := api.New(s, e)

	// 使用 gin.New() 代替 gin.Default()，避免默认 logger 和 recovery 的冗余日志
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())
	r.Use(slogMiddleware())
	a.RegisterRoutes(r)

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	slog.Info("MCPFlow server starting", "addr", addr)
	if err := r.Run(addr); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
