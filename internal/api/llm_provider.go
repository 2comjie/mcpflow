package api

import (
	"github.com/2comjie/mcpflow/internal/model"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (a *API) CreateLLMProvider(c *gin.Context) {
	var p model.LLMProvider
	if err := c.ShouldBindJSON(&p); err != nil {
		fail(c, 400, err.Error())
		return
	}
	if err := a.store.CreateLLMProvider(&p); err != nil {
		fail(c, 500, err.Error())
		return
	}
	ok(c, p)
}

func (a *API) GetLLMProvider(c *gin.Context) {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		fail(c, 400, "invalid id")
		return
	}
	p, err := a.store.GetLLMProvider(id)
	if err != nil {
		fail(c, 404, "llm provider not found")
		return
	}
	ok(c, p)
}

func (a *API) ListLLMProviders(c *gin.Context) {
	providers, err := a.store.ListLLMProviders()
	if err != nil {
		fail(c, 500, err.Error())
		return
	}
	ok(c, providers)
}

func (a *API) UpdateLLMProvider(c *gin.Context) {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		fail(c, 400, "invalid id")
		return
	}
	var updates map[string]any
	if err := c.ShouldBindJSON(&updates); err != nil {
		fail(c, 400, err.Error())
		return
	}
	delete(updates, "id")
	delete(updates, "_id")
	if err := a.store.UpdateLLMProvider(id, updates); err != nil {
		fail(c, 500, err.Error())
		return
	}
	p, _ := a.store.GetLLMProvider(id)
	ok(c, p)
}

func (a *API) DeleteLLMProvider(c *gin.Context) {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		fail(c, 400, "invalid id")
		return
	}
	if err := a.store.DeleteLLMProvider(id); err != nil {
		fail(c, 500, err.Error())
		return
	}
	ok(c, nil)
}
