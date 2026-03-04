package model

import "time"

type MCPServer struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	Name        string     `json:"name" gorm:"size:255;not null"`
	Description string     `json:"description" gorm:"type:text"`
	URL         string     `json:"url" gorm:"size:512;not null"`
	Headers     JSON       `json:"headers" gorm:"type:json"`
	Status      string     `json:"status" gorm:"size:20;default:inactive"`
	Tools       JSON       `json:"tools" gorm:"type:json"`
	Prompts     JSON       `json:"prompts" gorm:"type:json"`
	Resources   JSON       `json:"resources" gorm:"type:json"`
	CheckedAt   *time.Time `json:"checked_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (MCPServer) TableName() string { return "mcp_servers" }

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
