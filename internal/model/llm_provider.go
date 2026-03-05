package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type LLMProvider struct {
	ID        bson.ObjectID `json:"id" bson:"_id,omitempty"`
	Name      string        `json:"name" bson:"name"`
	BaseURL   string        `json:"base_url" bson:"base_url"`
	APIKey    string        `json:"api_key" bson:"api_key"`
	Models    []string      `json:"models" bson:"models"`
	CreatedAt time.Time     `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time     `json:"updated_at" bson:"updated_at"`
}
