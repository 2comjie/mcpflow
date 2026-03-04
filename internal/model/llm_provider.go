package model

import (
	"time"

	"github.com/2comjie/mcpflow/pkg/types"
)

type LLMProvider struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"size:255;not null"`
	BaseURL   string         `json:"base_url" gorm:"size:512;not null"`
	APIKey    string         `json:"api_key" gorm:"size:512;not null"`
	Models    types.JSONRaw  `json:"models" gorm:"type:json"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

func (LLMProvider) TableName() string { return "llm_providers" }
