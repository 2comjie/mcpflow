package store

import (
	"encoding/json"

	"github.com/2comjie/mcpflow/internal/model"
	"github.com/2comjie/mcpflow/pkg/types"
	"gorm.io/gorm"
)

// marshalJSONFields 自动将 map 中指定 key 的复杂值序列化为 JSONRaw。
// GORM 的 Updates(map[string]any) 不调用自定义类型的 Value() 方法，
// 所以需要提前将 map/slice 等复杂类型序列化为 JSONRaw。
func marshalJSONFields(updates map[string]any, keys ...string) {
	for _, k := range keys {
		if v, ok := updates[k]; ok {
			switch v.(type) {
			case string, int, int64, float64, bool, nil,
				types.JSONRaw, json.RawMessage:
			default:
				updates[k] = types.MustJSONRaw(v)
			}
		}
	}
}

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
