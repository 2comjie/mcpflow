package mcpserver

import (
	"time"

	"github.com/2comjie/mcpflow/pkg/types"
)

// MCPServer MCP服务器注册信息
type MCPServer struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:255;not null"`
	Description string    `json:"description" gorm:"type:text"`
	URL         string    `json:"url" gorm:"size:512;not null"`
	Status      string    `json:"status" gorm:"size:20;default:active"` // active/inactive
	Tools       JSON      `json:"tools" gorm:"type:json"`               // 缓存的工具列表
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (MCPServer) TableName() string {
	return "mcp_servers"
}

type JSON = types.JSONMap
