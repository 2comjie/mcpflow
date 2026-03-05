package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/2comjie/mcpflow/internal/engine"
	"github.com/2comjie/mcpflow/internal/model"
	"github.com/gin-gonic/gin"
)

func (a *API) CreateWorkflow(c *gin.Context) {
	var wf model.Workflow
	if err := c.ShouldBindJSON(&wf); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := a.store.CreateWorkflow(&wf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, wf)
}

func (a *API) GetWorkflow(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	wf, err := a.store.GetWorkflow(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workflow not found"})
		return
	}
	c.JSON(http.StatusOK, wf)
}

func (a *API) ListWorkflows(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	list, total, err := a.store.ListWorkflows(page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "total": total})
}

func (a *API) UpdateWorkflow(c *gin.Context) {
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
	if err := a.store.UpdateWorkflow(uint(id), updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (a *API) DeleteWorkflow(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := a.store.DeleteWorkflow(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (a *API) ExecuteWorkflow(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	wf, err := a.store.GetWorkflow(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workflow not found"})
		return
	}

	var input map[string]any
	if err := c.ShouldBindJSON(&input); err != nil {
		input = map[string]any{}
	}

	// 创建执行记录
	now := time.Now()
	exec := &model.Execution{
		WorkflowID: wf.ID,
		Status:     model.ExecPending,
		Input:      input,
		StartedAt:  &now,
	}
	if err := a.store.CreateExecution(exec); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 后台执行
	go func() {
		if err := a.store.UpdateExecution(exec.ID, map[string]any{"status": model.ExecRunning}); err != nil {
			log.Printf("update execution %d to running failed: %v", exec.ID, err)
		}

		logFn := func(el *model.ExecutionLog) {
			el.ExecutionID = exec.ID
			if err := a.store.CreateLog(el); err != nil {
				log.Printf("create execution log for node %s failed: %v", el.NodeID, err)
			}
		}

		// 使用独立 context，避免 HTTP 请求结束后 context 被取消
		ctx := context.Background()
		result, err := a.engine.Run(ctx, wf, input, logFn)

		finished := time.Now()
		updates := map[string]any{"finished_at": finished}

		if err != nil {
			updates["status"] = model.ExecFailed
			updates["error"] = err.Error()
		} else {
			updates["status"] = model.ExecCompleted
			updates["output"] = result.Output
			updates["node_states"] = result.NodeStates
		}

		if err := a.store.UpdateExecution(exec.ID, updates); err != nil {
			log.Printf("update execution %d failed: %v", exec.ID, err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"execution_id": exec.ID})
}

func (a *API) ListWorkflowExecutions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	list, total, err := a.store.ListExecutions(uint(id), page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "total": total})
}

// ExecuteWorkflowSSE 流式执行工作流，通过 SSE 推送节点事件
func (a *API) ExecuteWorkflowSSE(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	wf, err := a.store.GetWorkflow(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workflow not found"})
		return
	}

	var input map[string]any
	if err := c.ShouldBindJSON(&input); err != nil {
		input = map[string]any{}
	}

	// 创建执行记录
	now := time.Now()
	exec := &model.Execution{
		WorkflowID: wf.ID,
		Status:     model.ExecPending,
		Input:      input,
		StartedAt:  &now,
	}
	if err := a.store.CreateExecution(exec); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Flush()

	// 发送 execution_id 事件
	sendSSE(c, "execution_id", map[string]any{"execution_id": exec.ID})

	a.store.UpdateExecution(exec.ID, map[string]any{"status": model.ExecRunning})

	logFn := func(el *model.ExecutionLog) {
		el.ExecutionID = exec.ID
		if err := a.store.CreateLog(el); err != nil {
			log.Printf("create execution log for node %s failed: %v", el.NodeID, err)
		}
	}

	eventFn := func(event *engine.NodeEvent) {
		sendSSE(c, "node_event", event)
	}

	ctx := context.Background()
	result, execErr := a.engine.RunWithEvents(ctx, wf, input, logFn, eventFn)

	finished := time.Now()
	updates := map[string]any{"finished_at": finished}

	if execErr != nil {
		updates["status"] = model.ExecFailed
		updates["error"] = execErr.Error()
		sendSSE(c, "error", map[string]any{"error": execErr.Error()})
	} else {
		updates["status"] = model.ExecCompleted
		updates["output"] = result.Output
		updates["node_states"] = result.NodeStates
		sendSSE(c, "completed", map[string]any{"output": result.Output})
	}

	a.store.UpdateExecution(exec.ID, updates)

	// 发送 done 事件关闭流
	sendSSE(c, "done", map[string]any{"execution_id": exec.ID})
}

func sendSSE(c *gin.Context, event string, data any) {
	b, _ := json.Marshal(data)
	fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, string(b))
	c.Writer.Flush()
}
