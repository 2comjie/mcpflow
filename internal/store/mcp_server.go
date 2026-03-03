package store

import (
	"github.com/2comjie/mcpflow/internal/model"
	"gorm.io/gorm"
)

func (s *Store) CreateMCPServer(srv *model.MCPServer) error {
	return s.db.Create(srv).Error
}

func (s *Store) GetMCPServer(id uint) (*model.MCPServer, error) {
	var srv model.MCPServer
	if err := s.db.First(&srv, id).Error; err != nil {
		return nil, err
	}
	return &srv, nil
}

func (s *Store) ListMCPServers() ([]model.MCPServer, error) {
	var list []model.MCPServer
	if err := s.db.Order("id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (s *Store) UpdateMCPServer(id uint, updates map[string]any) error {
	return s.db.Model(&model.MCPServer{}).Where("id = ?", id).Updates(updates).Error
}

func (s *Store) DeleteMCPServer(id uint) error {
	return s.db.Delete(&model.MCPServer{}, id).Error
}

func (s *Store) UpdateMCPServerCache(id uint, updates map[string]any) error {
	return s.db.Model(&model.MCPServer{}).Where("id = ?", id).
		Select(keysOf(updates)).Updates(updates).Error
}

func keysOf(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// UpdateMCPServerFull 用结构体整体更新
func (s *Store) UpdateMCPServerFull(srv *model.MCPServer) error {
	return s.db.Session(&gorm.Session{FullSaveAssociations: true}).Save(srv).Error
}
