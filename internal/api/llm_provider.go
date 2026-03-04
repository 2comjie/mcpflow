package api

import (
	"net/http"
	"strconv"

	"github.com/2comjie/mcpflow/internal/model"
	"github.com/2comjie/mcpflow/pkg/types"
	"github.com/gin-gonic/gin"
)

func (a *API) CreateLLMProvider(c *gin.Context) {
	var p model.LLMProvider
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := a.store.CreateLLMProvider(&p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

func (a *API) ListLLMProviders(c *gin.Context) {
	list, err := a.store.ListLLMProviders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// API Key 脱敏
	for i := range list {
		list[i].APIKey = maskAPIKey(list[i].APIKey)
	}
	c.JSON(http.StatusOK, list)
}

func (a *API) UpdateLLMProvider(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var updates map[string]any
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if v, ok := updates["models"]; ok {
		updates["models"] = types.MustJSONRaw(v)
	}
	if err := a.store.UpdateLLMProvider(uint(id), updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (a *API) DeleteLLMProvider(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := a.store.DeleteLLMProvider(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// maskAPIKey 脱敏 API Key：显示前4后4字符
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}
