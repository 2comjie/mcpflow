package store

import "github.com/2comjie/mcpflow/internal/model"

func (s *Store) CreateLLMProvider(p *model.LLMProvider) error {
	return s.db.Create(p).Error
}

func (s *Store) GetLLMProvider(id uint) (*model.LLMProvider, error) {
	var p model.LLMProvider
	if err := s.db.First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) ListLLMProviders() ([]model.LLMProvider, error) {
	var list []model.LLMProvider
	if err := s.db.Order("id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (s *Store) UpdateLLMProvider(id uint, updates map[string]any) error {
	marshalJSONFields(updates, "models")
	return s.db.Model(&model.LLMProvider{}).Where("id = ?", id).Updates(updates).Error
}

func (s *Store) DeleteLLMProvider(id uint) error {
	return s.db.Delete(&model.LLMProvider{}, id).Error
}
