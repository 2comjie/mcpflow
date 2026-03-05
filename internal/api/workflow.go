package api

import (
	"github.com/2comjie/mcpflow/internal/model"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (a *API) CreateWorkflow(c *gin.Context) {
	var wf model.Workflow
	if err := c.ShouldBindJSON(&wf); err != nil {
		fail(c, 400, err.Error())
		return
	}
	if err := a.store.CreateWorkflow(&wf); err != nil {
		fail(c, 500, err.Error())
		return
	}
	ok(c, wf)
}

func (a *API) GetWorkflow(c *gin.Context) {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		fail(c, 400, "invalid id")
		return
	}
	wf, err := a.store.GetWorkflow(id)
	if err != nil {
		fail(c, 404, "workflow not found")
		return
	}
	ok(c, wf)
}

func (a *API) ListWorkflows(c *gin.Context) {
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "page_size", 20)
	workflows, total, err := a.store.ListWorkflows(page, pageSize)
	if err != nil {
		fail(c, 500, err.Error())
		return
	}
	okList(c, workflows, total)
}

func (a *API) UpdateWorkflow(c *gin.Context) {
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
	if err := a.store.UpdateWorkflow(id, updates); err != nil {
		fail(c, 500, err.Error())
		return
	}
	wf, _ := a.store.GetWorkflow(id)
	ok(c, wf)
}

func (a *API) DeleteWorkflow(c *gin.Context) {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		fail(c, 400, "invalid id")
		return
	}
	if err := a.store.DeleteWorkflow(id); err != nil {
		fail(c, 500, err.Error())
		return
	}
	ok(c, nil)
}

func (a *API) ExecuteWorkflow(c *gin.Context) {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		fail(c, 400, "invalid id")
		return
	}
	wf, err := a.store.GetWorkflow(id)
	if err != nil {
		fail(c, 404, "workflow not found")
		return
	}

	var input map[string]any
	_ = c.ShouldBindJSON(&input)
	if input == nil {
		input = make(map[string]any)
	}

	exec, err := a.engine.ExecuteWorkflow(wf, input)
	if err != nil {
		// 即使失败也返回执行记录（包含错误信息和节点状态）
		if exec != nil {
			c.JSON(200, gin.H{"data": exec, "error": err.Error()})
			return
		}
		fail(c, 500, err.Error())
		return
	}
	ok(c, exec)
}
