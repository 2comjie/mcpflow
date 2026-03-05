package store

import (
	"context"
	"time"

	"github.com/2comjie/mcpflow/internal/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (s *Store) CreateWorkflow(w *model.Workflow) error {
	now := time.Now()
	w.CreatedAt = now
	w.UpdatedAt = now
	result, err := s.workflows().InsertOne(context.TODO(), w)
	if err != nil {
		return err
	}
	w.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

func (s *Store) GetWorkflow(id bson.ObjectID) (*model.Workflow, error) {
	var w model.Workflow
	err := s.workflows().FindOne(context.TODO(), bson.M{"_id": id}).Decode(&w)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (s *Store) ListWorkflows(page, pageSize int) ([]model.Workflow, int64, error) {
	ctx := context.TODO()
	total, err := s.workflows().CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}

	skip := int64((page - 1) * pageSize)
	opts := options.Find().SetSkip(skip).SetLimit(int64(pageSize)).SetSort(bson.M{"created_at": -1})
	cursor, err := s.workflows().Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var workflows []model.Workflow
	if err := cursor.All(ctx, &workflows); err != nil {
		return nil, 0, err
	}
	return workflows, total, nil
}

func (s *Store) UpdateWorkflow(id bson.ObjectID, updates map[string]any) error {
	updates["updated_at"] = time.Now()
	_, err := s.workflows().UpdateByID(context.TODO(), id, bson.M{"$set": updates})
	return err
}

func (s *Store) DeleteWorkflow(id bson.ObjectID) error {
	_, err := s.workflows().DeleteOne(context.TODO(), bson.M{"_id": id})
	return err
}
