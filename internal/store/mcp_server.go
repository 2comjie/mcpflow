package store

import (
	"context"
	"time"

	"github.com/2comjie/mcpflow/internal/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (s *Store) CreateMCPServer(srv *model.MCPServer) error {
	now := time.Now()
	srv.CreatedAt = now
	srv.UpdatedAt = now
	if srv.Status == "" {
		srv.Status = "unknown"
	}
	result, err := s.mcpServers().InsertOne(context.TODO(), srv)
	if err != nil {
		return err
	}
	srv.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

func (s *Store) GetMCPServer(id bson.ObjectID) (*model.MCPServer, error) {
	var srv model.MCPServer
	err := s.mcpServers().FindOne(context.TODO(), bson.M{"_id": id}).Decode(&srv)
	if err != nil {
		return nil, err
	}
	return &srv, nil
}

func (s *Store) ListMCPServers() ([]model.MCPServer, error) {
	ctx := context.TODO()
	opts := options.Find().SetSort(bson.M{"created_at": -1})
	cursor, err := s.mcpServers().Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var servers []model.MCPServer
	if err := cursor.All(ctx, &servers); err != nil {
		return nil, err
	}
	return servers, nil
}

func (s *Store) UpdateMCPServer(id bson.ObjectID, updates map[string]any) error {
	updates["updated_at"] = time.Now()
	_, err := s.mcpServers().UpdateByID(context.TODO(), id, bson.M{"$set": updates})
	return err
}

func (s *Store) ReplaceMCPServer(srv *model.MCPServer) error {
	srv.UpdatedAt = time.Now()
	_, err := s.mcpServers().ReplaceOne(context.TODO(), bson.M{"_id": srv.ID}, srv)
	return err
}

func (s *Store) DeleteMCPServer(id bson.ObjectID) error {
	_, err := s.mcpServers().DeleteOne(context.TODO(), bson.M{"_id": id})
	return err
}
