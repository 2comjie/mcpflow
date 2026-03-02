package mcpserver

import (
	"context"
	"fmt"

	"github.com/2comjie/mcpflow/internal/mcp"
	"gorm.io/gorm"
)

type Service struct {
	db        *gorm.DB
	mcpClient *mcp.Client
}

func NewService(db *gorm.DB, mcpClient *mcp.Client) *Service {
	return &Service{db: db, mcpClient: mcpClient}
}

func (s *Service) AutoMigrate() error {
	return s.db.AutoMigrate(&MCPServer{})
}

func (s *Service) Create(ctx context.Context, server *MCPServer) error {
	if server.Name == "" || server.URL == "" {
		return fmt.Errorf("name and url are required")
	}
	server.Status = "active"
	return s.db.WithContext(ctx).Create(server).Error
}

func (s *Service) List(ctx context.Context) ([]MCPServer, error) {
	var servers []MCPServer
	err := s.db.WithContext(ctx).Order("id DESC").Find(&servers).Error
	return servers, err
}

func (s *Service) GetByID(ctx context.Context, id uint) (*MCPServer, error) {
	var server MCPServer
	err := s.db.WithContext(ctx).First(&server, id).Error
	if err != nil {
		return nil, err
	}
	return &server, nil
}

func (s *Service) Update(ctx context.Context, server *MCPServer) error {
	return s.db.WithContext(ctx).Model(server).Select("name", "description", "url").Updates(server).Error
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Delete(&MCPServer{}, id).Error
}

// TestConnection 测试连接并获取工具列表
func (s *Service) TestConnection(ctx context.Context, id uint) ([]map[string]any, error) {
	server, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	tools, err := s.mcpClient.ListTools(ctx, server.URL)
	if err != nil {
		// 标记为 inactive
		s.db.WithContext(ctx).Model(server).Update("status", "inactive")
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	// 连接成功，缓存工具列表，标记 active
	s.db.WithContext(ctx).Model(server).Updates(map[string]any{
		"status": "active",
		"tools":  map[string]any{"tools": tools},
	})

	return tools, nil
}

// GetTools 获取某个 MCP Server 的工具列表（优先用缓存）
func (s *Service) GetTools(ctx context.Context, id uint) ([]map[string]any, error) {
	server, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 尝试从缓存获取
	if server.Tools != nil {
		if tools, ok := server.Tools["tools"]; ok {
			if list, ok := tools.([]any); ok {
				result := make([]map[string]any, 0, len(list))
				for _, item := range list {
					if m, ok := item.(map[string]any); ok {
						result = append(result, m)
					}
				}
				return result, nil
			}
		}
	}

	// 缓存没有就实时获取
	return s.TestConnection(ctx, id)
}

// GetPrompts 获取某个 MCP Server 的提示词列表
func (s *Service) GetPrompts(ctx context.Context, id uint) ([]map[string]any, error) {
	server, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.mcpClient.ListPrompts(ctx, server.URL)
}

// GetResources 获取某个 MCP Server 的资源列表
func (s *Service) GetResources(ctx context.Context, id uint) ([]map[string]any, error) {
	server, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.mcpClient.ListResources(ctx, server.URL)
}
