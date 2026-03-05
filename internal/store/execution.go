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

func (s *Store) ListExecutions(workflowID bson.ObjectID, page, pageSize int) ([]model.Execution, int64, error) {
	ctx := context.TODO()
	filter := bson.M{"workflow_id": workflowID}

	total, err := s.executions().CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)
	opts := options.Find().SetSort(bson.M{"_id": -1}).SetSkip(skip).SetLimit(limit)

	cursor, err := s.executions().Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}

	var executions []model.Execution
	if err := cursor.All(ctx, &executions); err != nil {
		return nil, 0, err
	}
	if executions == nil {
		executions = []model.Execution{}
	}
	return executions, total, nil
}

func (s *Store) ListAllExecutions(page, pageSize int) ([]model.Execution, int64, error) {
	ctx := context.TODO()

	total, err := s.executions().CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}

	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)
	opts := options.Find().SetSort(bson.M{"_id": -1}).SetSkip(skip).SetLimit(limit)

	cursor, err := s.executions().Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, err
	}

	var executions []model.Execution
	if err := cursor.All(ctx, &executions); err != nil {
		return nil, 0, err
	}
	if executions == nil {
		executions = []model.Execution{}
	}
	return executions, total, nil
}

func (s *Store) DeleteExecution(id bson.ObjectID) error {
	ctx := context.TODO()
	// 先删关联的 logs
	_, _ = s.executionLogs().DeleteMany(ctx, bson.M{"execution_id": id})
	_, err := s.executions().DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (s *Store) GetExecutionStats() (map[string]any, error) {
	ctx := context.TODO()
	stats := map[string]any{}

	totalExec, _ := s.executions().CountDocuments(ctx, bson.M{})
	stats["total_executions"] = totalExec

	successExec, _ := s.executions().CountDocuments(ctx, bson.M{"status": model.ExecCompleted})
	stats["success_count"] = successExec

	failedExec, _ := s.executions().CountDocuments(ctx, bson.M{"status": model.ExecFailed})
	stats["failed_count"] = failedExec

	if totalExec > 0 {
		stats["success_rate"] = float64(successExec) / float64(totalExec) * 100
	} else {
		stats["success_rate"] = 0.0
	}

	// 计算平均耗时
	pipeline := bson.A{
		bson.M{"$match": bson.M{"status": "completed"}},
		bson.M{"$group": bson.M{"_id": nil, "avg_duration": bson.M{"$avg": "$duration"}}},
	}
	cursor, err := s.executionLogs().Aggregate(ctx, pipeline)
	if err == nil {
		var results []bson.M
		if cursor.All(ctx, &results) == nil && len(results) > 0 {
			if avg, ok := results[0]["avg_duration"]; ok {
				stats["avg_duration_ms"] = avg
			}
		}
	}
	if _, ok := stats["avg_duration_ms"]; !ok {
		stats["avg_duration_ms"] = 0.0
	}

	// 最近 5 条执行记录
	recentOpts := options.Find().SetSort(bson.M{"_id": -1}).SetLimit(5)
	recentCursor, err := s.executions().Find(ctx, bson.M{}, recentOpts)
	if err == nil {
		var recent []model.Execution
		if recentCursor.All(ctx, &recent) == nil {
			stats["recent_executions"] = recent
		}
	}
	if _, ok := stats["recent_executions"]; !ok {
		stats["recent_executions"] = []model.Execution{}
	}

	return stats, nil
}

func (s *Store) CreateLog(log *model.ExecutionLog) error {
	log.CreatedAt = time.Now()
	result, err := s.executionLogs().InsertOne(context.TODO(), log)
	if err != nil {
		return err
	}
	log.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

func (s *Store) ListLogs(executionID bson.ObjectID) ([]model.ExecutionLog, error) {
	ctx := context.TODO()
	opts := options.Find().SetSort(bson.M{"_id": 1})
	cursor, err := s.executionLogs().Find(ctx, bson.M{"execution_id": executionID}, opts)
	if err != nil {
		return nil, err
	}

	var logs []model.ExecutionLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, err
	}
	if logs == nil {
		logs = []model.ExecutionLog{}
	}
	return logs, nil
}
