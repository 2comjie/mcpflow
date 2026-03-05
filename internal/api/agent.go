package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/2comjie/mcpflow/internal/engine"
	"github.com/2comjie/mcpflow/internal/model"
	"github.com/gin-gonic/gin"
)

type AgentChatRequest struct {
	LLMProviderID uint   `json:"llm_provider_id" binding:"required"`
	MCPServerIDs  []uint `json:"mcp_server_ids" binding:"required"`
	Message       string `json:"message" binding:"required"`
	SystemMsg     string `json:"system_msg"`
	MaxIterations int    `json:"max_iterations"`
	Temperature   float64 `json:"temperature"`
	MaxTokens     int    `json:"max_tokens"`
}

func (a *API) AgentChat(c *gin.Context) {
	var req AgentChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取 LLM Provider
	provider, err := a.store.GetLLMProvider(req.LLMProviderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "llm provider not found"})
		return
	}

	// 获取 MCP Servers
	var mcpServers []model.AgentMCPServer
	for _, id := range req.MCPServerIDs {
		srv, err := a.store.GetMCPServer(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "mcp server not found: " + err.Error()})
			return
		}
		mcpServers = append(mcpServers, model.AgentMCPServer{
			URL:     srv.URL,
			Headers: srv.GetHeadersMap(),
		})
	}

	// 解析模型列表，取第一个可用模型
	modelName := ""
	if len(provider.Models) > 0 {
		var models []string
		if err := json.Unmarshal(provider.Models, &models); err == nil && len(models) > 0 {
			modelName = models[0]
		}
	}

	// 构建 Agent 配置
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

	// 执行 Agent
	executor := engine.NewAgentExecutor(a.mcp)
	result, err := executor.Execute(context.Background(), node, map[string]any{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
