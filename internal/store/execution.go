package store

import (
	"context"
	"time"

	"github.com/2comjie/mcpflow/internal/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (s *Store) CreateExecution(e *model.Execution) error {
	e.CreatedAt = time.Now()
	result, err := s.executions().InsertOne(context.TODO(), e)
	if err != nil {
		return err
	}
	e.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

func (s *Store) GetExecution(id bson.ObjectID) (*model.Execution, error) {
	var e model.Execution
	err := s.executions().FindOne(context.TODO(), bson.M{"_id": id}).Decode(&e)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (s *Store) UpdateExecution(id bson.ObjectID, updates map[string]any) error {
	_, err := s.executions().UpdateByID(context.TODO(), id, bson.M{"$set": updates})
	return err
}

func (s *Store) ListExecutions(page, pageSize int) ([]model.Execution, int64, error) {
	ctx := context.TODO()
	total, err := s.executions().CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}

	skip := int64((page - 1) * pageSize)
	opts := options.Find().SetSkip(skip).SetLimit(int64(pageSize)).SetSort(bson.M{"created_at": -1})
	cursor, err := s.executions().Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var execs []model.Execution
	if err := cursor.All(ctx, &execs); err != nil {
		return nil, 0, err
	}
	return execs, total, nil
}

func (s *Store) ListExecutionsByWorkflow(workflowID bson.ObjectID, page, pageSize int) ([]model.Execution, int64, error) {
	ctx := context.TODO()
	filter := bson.M{"workflow_id": workflowID}
	total, err := s.executions().CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	skip := int64((page - 1) * pageSize)
	opts := options.Find().SetSkip(skip).SetLimit(int64(pageSize)).SetSort(bson.M{"created_at": -1})
	cursor, err := s.executions().Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var execs []model.Execution
	if err := cursor.All(ctx, &execs); err != nil {
		return nil, 0, err
	}
	return execs, total, nil
}

func (s *Store) DeleteExecution(id bson.ObjectID) error {
	ctx := context.TODO()
	_, err := s.executions().DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	_, err = s.executionLogs().DeleteMany(ctx, bson.M{"execution_id": id})
	return err
}

// ==================== Execution Logs ====================

func (s *Store) CreateExecutionLog(log *model.ExecutionLog) error {
	log.CreatedAt = time.Now()
	result, err := s.executionLogs().InsertOne(context.TODO(), log)
	if err != nil {
		return err
	}
	log.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

func (s *Store) GetExecutionLogs(executionID bson.ObjectID) ([]model.ExecutionLog, error) {
	ctx := context.TODO()
	opts := options.Find().SetSort(bson.M{"created_at": 1})
	cursor, err := s.executionLogs().Find(ctx, bson.M{"execution_id": executionID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []model.ExecutionLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, err
	}
	return logs, nil
}

// ==================== Stats ====================

func (s *Store) GetExecutionStats() (map[string]any, error) {
	ctx := context.TODO()
	total, _ := s.executions().CountDocuments(ctx, bson.M{})
	success, _ := s.executions().CountDocuments(ctx, bson.M{"status": "completed"})
	failed, _ := s.executions().CountDocuments(ctx, bson.M{"status": "failed"})

	stats := map[string]any{
		"total_executions": total,
		"success_count":    success,
		"failed_count":     failed,
	}

	if total > 0 {
		stats["success_rate"] = float64(success) / float64(total) * 100
	} else {
		stats["success_rate"] = float64(0)
	}

	wfCount, _ := s.workflows().CountDocuments(ctx, bson.M{})
	stats["total_workflows"] = wfCount

	return stats, nil
}
