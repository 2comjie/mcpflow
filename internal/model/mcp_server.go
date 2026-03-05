package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type MCPServer struct {
	ID          bson.ObjectID     `json:"id" bson:"_id,omitempty"`
	Name        string            `json:"name" bson:"name"`
	Description string            `json:"description" bson:"description"`
	URL         string            `json:"url" bson:"url"`
	Headers     map[string]string `json:"headers" bson:"headers"`
	Status      string            `json:"status" bson:"status"`
	Tools       any               `json:"tools" bson:"tools"`
	Prompts     any               `json:"prompts" bson:"prompts"`
	Resources   any               `json:"resources" bson:"resources"`
	CheckedAt   *time.Time        `json:"checked_at" bson:"checked_at"`
	CreatedAt   time.Time         `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" bson:"updated_at"`
}
