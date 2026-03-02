package workflow

import (
	"context"

	"gorm.io/gorm"
)

type WorkflowRepository struct {
	db *gorm.DB
}

func NewWorkflowRepository(db *gorm.DB) *WorkflowRepository {
	return &WorkflowRepository{db: db}
}

func (r *WorkflowRepository) AutoMigrate() error {
	// 自动建表
	return r.db.AutoMigrate(&Workflow{}, &WorkflowExecution{}, &ExecutionLog{})
}

// ==================== Workflow CRUD ====================
func (r *WorkflowRepository) Create(ctx context.Context, w *Workflow) error {
	return r.db.WithContext(ctx).Create(w).Error
}

func (r *WorkflowRepository) GetByID(ctx context.Context, id uint) (*Workflow, error) {
	var w Workflow
	err := r.db.WithContext(ctx).First(&w, id).Error
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *WorkflowRepository) List(ctx context.Context, offset, limit int) ([]Workflow, int64, error) {
	var workflows []Workflow
	var total int64

	db := r.db.WithContext(ctx).Model(&Workflow{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset(offset).Limit(limit).Order("id DESC").Find(&workflows).Error; err != nil {
		return nil, 0, err
	}
	return workflows, total, nil
}

func (r *WorkflowRepository) Update(ctx context.Context, w *Workflow) error {
	return r.db.WithContext(ctx).Model(w).
		Select("name", "description", "status", "nodes", "edges", "variables").
		Updates(w).Error
}

func (r *WorkflowRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Workflow{}, id).Error
}

// ==================== WorkflowExecution CRUD ====================

func (r *WorkflowRepository) CreateExecution(ctx context.Context, e *WorkflowExecution) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *WorkflowRepository) GetExecution(ctx context.Context, id uint) (*WorkflowExecution, error) {
	var e WorkflowExecution
	err := r.db.WithContext(ctx).First(&e, id).Error
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *WorkflowRepository) UpdateExecution(ctx context.Context, e *WorkflowExecution) error {
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *WorkflowRepository) ListExecutions(ctx context.Context, workflowID uint, offset, limit int) ([]WorkflowExecution, int64, error) {
	var executions []WorkflowExecution
	var total int64

	db := r.db.WithContext(ctx).Model(&WorkflowExecution{}).Where("workflow_id = ?", workflowID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset(offset).Limit(limit).Order("id DESC").Find(&executions).Error; err != nil {
		return nil, 0, err
	}
	return executions, total, nil
}

func (r *WorkflowRepository) ListAllExecutions(ctx context.Context, offset, limit int) ([]WorkflowExecution, int64, error) {
	var executions []WorkflowExecution
	var total int64

	db := r.db.WithContext(ctx).Model(&WorkflowExecution{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset(offset).Limit(limit).Order("id DESC").Find(&executions).Error; err != nil {
		return nil, 0, err
	}
	return executions, total, nil
}

// ==================== ExecutionLog ====================

func (r *WorkflowRepository) CreateLog(ctx context.Context, log *ExecutionLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *WorkflowRepository) ListLogs(ctx context.Context, executionID uint) ([]ExecutionLog, error) {
	var logs []ExecutionLog
	err := r.db.WithContext(ctx).Where("execution_id = ?", executionID).Order("id ASC").Find(&logs).Error
	return logs, err
}
