package api

import (
	"context"
	"net/http"

	"github.com/2comjie/mcpflow/internal/engine"
	"github.com/2comjie/mcpflow/internal/model"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type AgentChatRequest struct {
	LLMProviderID string   `json:"llm_provider_id" binding:"required"`
	MCPServerIDs  []string `json:"mcp_server_ids" binding:"required"`
	Message       string   `json:"message" binding:"required"`
	SystemMsg     string   `json:"system_msg"`
	MaxIterations int      `json:"max_iterations"`
	Temperature   float64  `json:"temperature"`
	MaxTokens     int      `json:"max_tokens"`
}

func (a *API) AgentChat(c *gin.Context) {
	var req AgentChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	providerID, err := bson.ObjectIDFromHex(req.LLMProviderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid llm_provider_id"})
		return
	}
	provider, err := a.store.GetLLMProvider(providerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "llm provider not found"})
		return
	}

	var mcpServers []model.AgentMCPServer
	for _, idStr := range req.MCPServerIDs {
		srvID, err := bson.ObjectIDFromHex(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid mcp_server_id: " + idStr})
			return
		}
		srv, err := a.store.GetMCPServer(srvID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "mcp server not found: " + idStr})
			return
		}
		mcpServers = append(mcpServers, model.AgentMCPServer{
			URL:     srv.URL,
			Headers: srv.GetHeadersMap(),
		})
	}

	modelName := ""
	if len(provider.Models) > 0 {
		modelName = provider.Models[0]
	}

	agentCfg := &model.AgentConfig{
		BaseURL:       provider.BaseURL,
		APIKey:        provider.APIKey,
		Model:         modelName,
		Prompt:        req.Message,
		SystemMsg:     req.SystemMsg,
		McpServers:    mcpServers,
		MaxIterations: req.MaxIterations,
		Temperature:   req.Temperature,
		MaxTokens:     req.MaxTokens,
	}

	if agentCfg.MaxIterations <= 0 {
		agentCfg.MaxIterations = 10
	}
	if agentCfg.MaxTokens <= 0 {
		agentCfg.MaxTokens = 2048
	}

	node := &model.Node{
		ID:   "playground_agent",
		Type: model.NodeAgent,
		Name: "Playground Agent",
		Config: model.NodeConfig{
			Agent: agentCfg,
		},
	}

	executor := engine.NewAgentExecutor(a.mcp)
	result, err := executor.Execute(context.Background(), node, map[string]any{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
