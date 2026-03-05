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

func (s *Store) GetExecutionStats() (map[string]any, error) {
	stats := map[string]any{}

	var totalExec int64
	s.db.Model(&model.Execution{}).Count(&totalExec)
	stats["total_executions"] = totalExec

	var successExec int64
	s.db.Model(&model.Execution{}).Where("status = ?", model.ExecCompleted).Count(&successExec)
	stats["success_count"] = successExec

	var failedExec int64
	s.db.Model(&model.Execution{}).Where("status = ?", model.ExecFailed).Count(&failedExec)
	stats["failed_count"] = failedExec

	if totalExec > 0 {
		stats["success_rate"] = float64(successExec) / float64(totalExec) * 100
	} else {
		stats["success_rate"] = 0.0
	}

	var avgDuration float64
	s.db.Model(&model.ExecutionLog{}).Where("status = ?", "completed").Select("COALESCE(AVG(duration), 0)").Row().Scan(&avgDuration)
	stats["avg_duration_ms"] = avgDuration

	var recent []model.Execution
	s.db.Model(&model.Execution{}).Order("id DESC").Limit(5).Find(&recent)
	stats["recent_executions"] = recent

	return stats, nil
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
