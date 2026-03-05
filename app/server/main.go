package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

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
	e := engine.New(s)
	a := api.New(s, e)

	r := gin.Default()
	r.Use(corsMiddleware())
	a.RegisterRoutes(r)

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("MCPFlow server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
