package store

import (
	"github.com/2comjie/mcpflow/internal/model"
	"gorm.io/gorm"
)

func (s *Store) CreateExecution(e *model.Execution) error {
	return s.db.Create(e).Error
}

func (s *Store) GetExecution(id uint) (*model.Execution, error) {
	var e model.Execution
	if err := s.db.First(&e, id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (s *Store) UpdateExecution(id uint, updates map[string]any) error {
	marshalJSONFields(updates, "input", "output", "node_states")
	return s.db.Model(&model.Execution{}).Where("id = ?", id).Updates(updates).Error
}

func (s *Store) ListExecutions(workflowID uint, page, pageSize int) ([]model.Execution, int64, error) {
	var executions []model.Execution
	var total int64

	db := s.db.Model(&model.Execution{}).Where("workflow_id = ?", workflowID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("id DESC").Offset(offset).Limit(pageSize).Find(&executions).Error; err != nil {
		return nil, 0, err
	}
	return executions, total, nil
}

func (s *Store) ListAllExecutions(page, pageSize int) ([]model.Execution, int64, error) {
	var executions []model.Execution
	var total int64

	db := s.db.Model(&model.Execution{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("id DESC").Offset(offset).Limit(pageSize).Find(&executions).Error; err != nil {
		return nil, 0, err
	}
	return executions, total, nil
}

func (s *Store) DeleteExecution(id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("execution_id = ?", id).Delete(&model.ExecutionLog{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Execution{}, id).Error
	})
}

func (s *Store) CreateLog(log *model.ExecutionLog) error {
	return s.db.Create(log).Error
}

func (s *Store) ListLogs(executionID uint) ([]model.ExecutionLog, error) {
	var logs []model.ExecutionLog
	if err := s.db.Where("execution_id = ?", executionID).Order("id ASC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}
