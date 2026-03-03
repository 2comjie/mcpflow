package mcpserver

import (
	"time"

	"github.com/2comjie/mcpflow/pkg/types"
)

// MCPServer MCP服务器注册信息
type MCPServer struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	Name        string     `json:"name" gorm:"size:255;not null"`
	Description string     `json:"description" gorm:"type:text"`
	URL         string     `json:"url" gorm:"size:512;not null"`
	Headers     JSON       `json:"headers" gorm:"type:json"`              // 自定义HTTP请求头
	Status      string     `json:"status" gorm:"size:20;default:active"`  // active/inactive
	Tools       JSON       `json:"tools" gorm:"type:json"`                // 缓存的工具列表
	Prompts     JSON       `json:"prompts" gorm:"type:json"`              // 缓存的提示词列表
	Resources   JSON       `json:"resources" gorm:"type:json"`            // 缓存的资源列表
	CheckedAt   *time.Time `json:"checked_at"`                            // 最后健康检查时间
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// GetHeadersMap 将 JSON headers 转换为 map[string]string
func (s *MCPServer) GetHeadersMap() map[string]string {
	if s.Headers == nil {
		return nil
	}
	result := make(map[string]string, len(s.Headers))
	for k, v := range s.Headers {
		if str, ok := v.(string); ok {
			result[k] = str
		}
	}
	return result
}

func (MCPServer) TableName() string {
	return "mcp_servers"
}

type JSON = types.JSONMap
