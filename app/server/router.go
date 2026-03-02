package main

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, app *App) {
	api := r.Group("/api/v1")

	app.Workflow.RegisterRoutes(api)
	app.MCPServer.RegisterRoutes(api)
	app.Secret.RegisterRoutes(api)
}
