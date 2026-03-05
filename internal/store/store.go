package store

import "go.mongodb.org/mongo-driver/v2/mongo"

type Store struct {
	db *mongo.Database
}

func New(db *mongo.Database) *Store {
	return &Store{db: db}
}

func (s *Store) workflows() *mongo.Collection    { return s.db.Collection("workflows") }
func (s *Store) executions() *mongo.Collection    { return s.db.Collection("executions") }
func (s *Store) executionLogs() *mongo.Collection { return s.db.Collection("execution_logs") }
func (s *Store) mcpServers() *mongo.Collection    { return s.db.Collection("mcp_servers") }
func (s *Store) llmProviders() *mongo.Collection  { return s.db.Collection("llm_providers") }
