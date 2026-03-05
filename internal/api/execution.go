package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (a *API) ListExecutions(c *gin.Context) {
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "page_size", 20)
	execs, total, err := a.store.ListExecutions(page, pageSize)
	if err != nil {
		fail(c, 500, err.Error())
		return
	}
	okList(c, execs, total)
}

func (a *API) GetExecution(c *gin.Context) {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		fail(c, 400, "invalid id")
		return
	}
	exec, err := a.store.GetExecution(id)
	if err != nil {
		fail(c, 404, "execution not found")
		return
	}
	ok(c, exec)
}

func (a *API) GetExecutionLogs(c *gin.Context) {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		fail(c, 400, "invalid id")
		return
	}
	logs, err := a.store.GetExecutionLogs(id)
	if err != nil {
		fail(c, 500, err.Error())
		return
	}
	ok(c, logs)
}

func (a *API) DeleteExecution(c *gin.Context) {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		fail(c, 400, "invalid id")
		return
	}
	if err := a.store.DeleteExecution(id); err != nil {
		fail(c, 500, err.Error())
		return
	}
	ok(c, nil)
}

func (a *API) ListWorkflowExecutions(c *gin.Context) {
	wfID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, 400, "invalid id")
		return
	}
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "page_size", 20)
	execs, total, err := a.store.ListExecutionsByWorkflow(wfID, page, pageSize)
	if err != nil {
		fail(c, 500, err.Error())
		return
	}
	okList(c, execs, total)
}
