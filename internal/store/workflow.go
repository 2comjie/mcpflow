package store

import (
	"github.com/2comjie/mcpflow/internal/model"
)

func (s *Store) CreateWorkflow(w *model.Workflow) error {
	return s.db.Create(w).Error
}

func (s *Store) GetWorkflow(id uint) (*model.Workflow, error) {
	var w model.Workflow
	if err := s.db.First(&w, id).Error; err != nil {
		return nil, err
	}
	return &w, nil
}

func (s *Store) ListWorkflows(page, pageSize int) ([]model.Workflow, int64, error) {
	var workflows []model.Workflow
	var total int64

	db := s.db.Model(&model.Workflow{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("id DESC").Offset(offset).Limit(pageSize).Find(&workflows).Error; err != nil {
		return nil, 0, err
	}
	return workflows, total, nil
}

func (s *Store) UpdateWorkflow(id uint, updates map[string]any) error {
	return s.db.Model(&model.Workflow{}).Where("id = ?", id).Updates(updates).Error
}

func (s *Store) SaveWorkflow(w *model.Workflow) error {
	return s.db.Save(w).Error
}

func (s *Store) DeleteWorkflow(id uint) error {
	return s.db.Delete(&model.Workflow{}, id).Error
}
