package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type ExecStatus string

const (
	ExecPending   ExecStatus = "pending"
	ExecRunning   ExecStatus = "running"
	ExecCompleted ExecStatus = "completed"
	ExecFailed    ExecStatus = "failed"
)

type Execution struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	WorkflowID uint       `json:"workflow_id" gorm:"index;not null"`
	Status     ExecStatus `json:"status" gorm:"size:20;default:pending"`
	Input      JSON       `json:"input" gorm:"type:json"`
	Output     JSON       `json:"output" gorm:"type:json"`
	NodeStates NodeStates `json:"node_states" gorm:"type:json"`
	Error      string     `json:"error" gorm:"type:text"`
	StartedAt  *time.Time `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (Execution) TableName() string { return "executions" }

type ExecutionLog struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	ExecutionID uint      `json:"execution_id" gorm:"index;not null"`
	NodeID      string    `json:"node_id" gorm:"size:100;not null"`
	NodeName    string    `json:"node_name" gorm:"size:255"`
	NodeType    NodeType  `json:"node_type" gorm:"size:50"`
	Attempt     int       `json:"attempt"`
	Status      string    `json:"status" gorm:"size:20"`
	Input       JSON      `json:"input" gorm:"type:json"`
	Output      JSON      `json:"output" gorm:"type:json"`
	Error       string    `json:"error" gorm:"type:text"`
	Duration    int64     `json:"duration"`
	CreatedAt   time.Time `json:"created_at"`
}

func (ExecutionLog) TableName() string { return "execution_logs" }

type NodeState struct {
	NodeID   string `json:"node_id"`
	Status   string `json:"status"`
	Input    any    `json:"input,omitempty"`
	Output   any    `json:"output,omitempty"`
	Error    string `json:"error,omitempty"`
	Duration int64  `json:"duration"`
}

type NodeStates map[string]NodeState

func (ns NodeStates) Value() (driver.Value, error) {
	if ns == nil {
		return "{}", nil
	}
	b, err := json.Marshal(ns)
	return string(b), err
}

func (ns *NodeStates) Scan(value any) error {
	if value == nil {
		*ns = make(NodeStates)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan NodeStates: %v", value)
	}
	return json.Unmarshal(bytes, ns)
}
