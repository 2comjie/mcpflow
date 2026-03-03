package mcpserver

import (
	"context"
	"fmt"
	"time"

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
	return s.db.WithContext(ctx).Model(server).Select("name", "description", "url", "headers").Updates(server).Error
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Delete(&MCPServer{}, id).Error
}

func (s *Service) TestConnection(ctx context.Context, id uint) ([]map[string]any, error) {
	server, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	headers := server.GetHeadersMap()

	tools, err := s.mcpClient.ListTools(ctx, server.URL, headers)
	if err != nil {
		now := time.Now()
		s.db.WithContext(ctx).Model(server).Updates(map[string]any{
			"status":     "inactive",
			"checked_at": &now,
		})
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	// 连接成功，同时尝试获取 prompts 和 resources
	now := time.Now()
	updates := map[string]any{
		"status":     "active",
		"tools":      map[string]any{"tools": tools},
		"checked_at": &now,
	}

	if prompts, err := s.mcpClient.ListPrompts(ctx, server.URL, headers); err == nil {
		updates["prompts"] = map[string]any{"prompts": prompts}
	}
	if resources, err := s.mcpClient.ListResources(ctx, server.URL, headers); err == nil {
		updates["resources"] = map[string]any{"resources": resources}
	}

	s.db.WithContext(ctx).Model(server).Updates(updates)
	return tools, nil
}

func (s *Service) HealthCheck(ctx context.Context, id uint) error {
	server, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	headers := server.GetHeadersMap()

	now := time.Now()
	if err := s.mcpClient.Ping(ctx, server.URL, headers); err != nil {
		s.db.WithContext(ctx).Model(server).Updates(map[string]any{
			"status":     "inactive",
			"checked_at": &now,
		})
		return fmt.Errorf("health check failed: %w", err)
	}

	s.db.WithContext(ctx).Model(server).Updates(map[string]any{
		"status":     "active",
		"checked_at": &now,
	})
	return nil
}

func (s *Service) HealthCheckAll(ctx context.Context) ([]MCPServer, error) {
	servers, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	for i := range servers {
		srv := &servers[i]
		headers := srv.GetHeadersMap()
		now := time.Now()
		if err := s.mcpClient.Ping(ctx, srv.URL, headers); err != nil {
			srv.Status = "inactive"
		} else {
			srv.Status = "active"
		}
		srv.CheckedAt = &now
		s.db.WithContext(ctx).Model(srv).Updates(map[string]any{
			"status":     srv.Status,
			"checked_at": &now,
		})
	}
	return servers, nil
}

func (s *Service) GetTools(ctx context.Context, id uint) ([]map[string]any, error) {
	server, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 尝试从缓存获取
	if cached := extractCachedList(server.Tools, "tools"); cached != nil {
		return cached, nil
	}

	// 缓存没有就实时获取
	return s.TestConnection(ctx, id)
}

func (s *Service) GetPrompts(ctx context.Context, id uint) ([]map[string]any, error) {
	server, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if cached := extractCachedList(server.Prompts, "prompts"); cached != nil {
		return cached, nil
	}

	headers := server.GetHeadersMap()
	prompts, err := s.mcpClient.ListPrompts(ctx, server.URL, headers)
	if err != nil {
		return nil, err
	}
	// 缓存结果
	s.db.WithContext(ctx).Model(server).Update("prompts", map[string]any{"prompts": prompts})
	return prompts, nil
}

func (s *Service) GetResources(ctx context.Context, id uint) ([]map[string]any, error) {
	server, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if cached := extractCachedList(server.Resources, "resources"); cached != nil {
		return cached, nil
	}

	headers := server.GetHeadersMap()
	resources, err := s.mcpClient.ListResources(ctx, server.URL, headers)
	if err != nil {
		return nil, err
	}
	// 缓存结果
	s.db.WithContext(ctx).Model(server).Update("resources", map[string]any{"resources": resources})
	return resources, nil
}

func extractCachedList(cache map[string]any, key string) []map[string]any {
	if cache == nil {
		return nil
	}
	items, ok := cache[key]
	if !ok {
		return nil
	}
	list, ok := items.([]any)
	if !ok {
		return nil
	}
	result := make([]map[string]any, 0, len(list))
	for _, item := range list {
		if m, ok := item.(map[string]any); ok {
			result = append(result, m)
		}
	}
	return result
}
