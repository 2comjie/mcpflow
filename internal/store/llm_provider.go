package store

import (
	"context"
	"time"

	"github.com/2comjie/mcpflow/internal/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (s *Store) CreateLLMProvider(p *model.LLMProvider) error {
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now
	result, err := s.llmProviders().InsertOne(context.TODO(), p)
	if err != nil {
		return err
	}
	p.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

func (s *Store) GetLLMProvider(id bson.ObjectID) (*model.LLMProvider, error) {
	var p model.LLMProvider
	err := s.llmProviders().FindOne(context.TODO(), bson.M{"_id": id}).Decode(&p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) ListLLMProviders() ([]model.LLMProvider, error) {
	ctx := context.TODO()
	opts := options.Find().SetSort(bson.M{"created_at": -1})
	cursor, err := s.llmProviders().Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var providers []model.LLMProvider
	if err := cursor.All(ctx, &providers); err != nil {
		return nil, err
	}
	if providers == nil {
		providers = []model.LLMProvider{}
	}
	return providers, nil
}

func (s *Store) UpdateLLMProvider(id bson.ObjectID, updates map[string]any) error {
	updates["updated_at"] = time.Now()
	_, err := s.llmProviders().UpdateByID(context.TODO(), id, bson.M{"$set": updates})
	return err
}

func (s *Store) DeleteLLMProvider(id bson.ObjectID) error {
	_, err := s.llmProviders().DeleteOne(context.TODO(), bson.M{"_id": id})
	return err
}
