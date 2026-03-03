package store

import (
	"github.com/2comjie/mcpflow/internal/model"
	"gorm.io/gorm"
)

type Store struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Store {
	return &Store{db: db}
}

func (s *Store) DB() *gorm.DB {
	return s.db
}

func (s *Store) AutoMigrate() error {
	return s.db.AutoMigrate(
		&model.Workflow{},
		&model.Execution{},
		&model.ExecutionLog{},
		&model.MCPServer{},
		&model.LLMProvider{},
	)
}
