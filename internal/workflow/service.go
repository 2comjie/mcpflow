package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type WorkflowService struct {
	repo    *WorkflowRepository
	engine  *Engine
	cancels sync.Map // executionID -> context.CancelFunc
}

func NewWorkflowService(repo *WorkflowRepository, engine *Engine) *WorkflowService {
	return &WorkflowService{repo: repo, engine: engine}
}

// ==================== 工作流 CRUD ====================

func (s *WorkflowService) Create(ctx context.Context, wf *Workflow) error {
	if err := validateWorkflow(wf); err != nil {
		return err
	}
	wf.Status = "draft"
	return s.repo.Create(ctx, wf)
}

func (s *WorkflowService) GetByID(ctx context.Context, id uint) (*Workflow, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *WorkflowService) List(ctx context.Context, page, pageSize int) ([]Workflow, int64, error) {
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}
	return s.repo.List(ctx, offset, pageSize)
}

func (s *WorkflowService) Update(ctx context.Context, wf *Workflow) error {
	if err := validateWorkflow(wf); err != nil {
		return err
	}
	return s.repo.Update(ctx, wf)
}

func (s *WorkflowService) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

// ==================== 执行工作流 ====================

func (s *WorkflowService) Execute(ctx context.Context, workflowID uint, input map[string]any) (*WorkflowExecution, error) {
	wf, err := s.repo.GetByID(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("workflow not found: %w", err)
	}

	now := time.Now()
	exec := &WorkflowExecution{
		WorkflowID: wf.ID,
		Status:     ExecRunning,
		Input:      JSON(input),
		StartedAt:  &now,
	}
	if err := s.repo.CreateExecution(ctx, exec); err != nil {
		return nil, err
	}

	// 执行上下文
	runCtx, cancel := context.WithCancel(ctx)
	s.cancels.Store(exec.ID, cancel)
	defer s.cancels.Delete(exec.ID)

	// 执行DAG
	logFn := func(l *ExecutionLog) {
		l.ExecutionID = exec.ID
		s.repo.CreateLog(ctx, l)
	}
	output, nodeStates, runErr := s.engine.Run(runCtx, wf, input, nil, logFn)

	// 更新执行结果
	finished := time.Now()
	exec.FinishedAt = &finished
	exec.NodeStates = NodeStates(nodeStates)

	if runErr != nil {
		exec.Status = ExecFailed
		exec.Error = runErr.Error()
	} else {
		exec.Status = ExecCompleted
		exec.Output = JSON(output)
	}

	if err := s.repo.UpdateExecution(ctx, exec); err != nil {
		return nil, err
	}

	return exec, runErr
}

func (s *WorkflowService) GetExecution(ctx context.Context, id uint) (*WorkflowExecution, error) {
	return s.repo.GetExecution(ctx, id)
}

func (s *WorkflowService) ListExecutions(ctx context.Context, workflowID uint, page, pageSize int) ([]WorkflowExecution, int64, error) {
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListExecutions(ctx, workflowID, offset, pageSize)
}

func (s *WorkflowService) GetExecutionLogs(ctx context.Context, executionID uint) ([]ExecutionLog, error) {
	return s.repo.ListLogs(ctx, executionID)
}

// ==================== 校验 ====================

func validateWorkflow(wf *Workflow) error {
	if wf.Name == "" {
		return fmt.Errorf("workflow name is required")
	}

	// 检查必须有 start 和 end 节点
	hasStart, hasEnd := false, false
	for _, n := range wf.Nodes {
		if n.Type == NodeStart {
			hasStart = true
		}
		if n.Type == NodeEnd {
			hasEnd = true
		}
	}
	if !hasStart {
		return fmt.Errorf("workflow must have a start node")
	}
	if !hasEnd {
		return fmt.Errorf("workflow must have an end node")
	}

	return nil
}

// 异步执行工作流，返回事件流
func (s *WorkflowService) ExecuteWithEvents(ctx context.Context, workflowID uint, input map[string]any) (*WorkflowExecution, *EventBus, error) {
	wf, err := s.GetByID(ctx, workflowID)
	if err != nil {
		return nil, nil, fmt.Errorf("workflow not found %w", err)
	}

	now := time.Now()
	exec := &WorkflowExecution{
		WorkflowID: wf.ID,
		Status:     ExecRunning,
		Input:      JSON(input),
		StartedAt:  &now,
	}
	if err := s.repo.CreateExecution(ctx, exec); err != nil {
		return nil, nil, err
	}

	eventBus := NewEventBus()
	// 执行上下文
	runCtx, cancel := context.WithCancel(ctx)
	s.cancels.Store(exec.ID, cancel)

	logFn := func(l *ExecutionLog) {
		l.ExecutionID = exec.ID
		s.repo.CreateLog(ctx, l)
	}

	// 异步执行
	go func() {
		defer s.cancels.Delete(exec.ID)
		output, nodeStates, runErr := s.engine.Run(runCtx, wf, input, eventBus, logFn)

		finished := time.Now()
		exec.FinishedAt = &finished
		exec.NodeStates = nodeStates
		if runErr != nil {
			exec.Status = ExecFailed
			exec.Error = runErr.Error()
		} else {
			exec.Status = ExecCompleted
			exec.Output = output
		}
		s.repo.UpdateExecution(ctx, exec)
	}()

	return exec, eventBus, nil
}

func (s *WorkflowService) CancelExecution(ctx context.Context, executionID uint) error {
	cancel, ok := s.cancels.Load(executionID)
	if !ok {
		return fmt.Errorf("execution %d is not running", executionID)
	}
	cancel.(context.CancelFunc)()

	exec, err := s.repo.GetExecution(ctx, executionID)
	if err != nil {
		return err
	}
	now := time.Now()
	exec.Status = ExecCancelled
	exec.FinishedAt = &now
	return s.repo.UpdateExecution(ctx, exec)
}
