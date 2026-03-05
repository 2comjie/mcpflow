package store

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

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
func (s *Store) counters() *mongo.Collection      { return s.db.Collection("counters") }

// NextSeqID 原子自增计数器，返回下一个序列号
func (s *Store) NextSeqID(name string) (int64, error) {
	var result struct {
		Seq int64 `bson:"seq"`
	}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	err := s.counters().FindOneAndUpdate(
		context.TODO(),
		bson.M{"_id": name},
		bson.M{"$inc": bson.M{"seq": int64(1)}},
		opts,
	).Decode(&result)
	if err != nil {
		return 0, err
	}
	return result.Seq, nil
}
