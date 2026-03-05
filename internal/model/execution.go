package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ExecStatus string

const (
	ExecPending   ExecStatus = "pending"
	ExecRunning   ExecStatus = "running"
	ExecCompleted ExecStatus = "completed"
	ExecFailed    ExecStatus = "failed"
)

type Execution struct {
	ID         bson.ObjectID        `json:"id" bson:"_id,omitempty"`
	WorkflowID int64                `json:"workflow_id" bson:"workflow_id"`
	Status     ExecStatus           `json:"status" bson:"status"`
	Input      map[string]any       `json:"input" bson:"input"`
	Output     map[string]any       `json:"output" bson:"output"`
	NodeStates map[string]NodeState `json:"node_states" bson:"node_states"`
	Error      string               `json:"error" bson:"error"`
	StartedAt  *time.Time           `json:"started_at" bson:"started_at"`
	FinishedAt *time.Time           `json:"finished_at" bson:"finished_at"`
	CreatedAt  time.Time            `json:"created_at" bson:"created_at"`
}

type NodeState struct {
	NodeID   string `json:"node_id" bson:"node_id"`
	Status   string `json:"status" bson:"status"`
	Input    any    `json:"input" bson:"input"`
	Output   any    `json:"output" bson:"output"`
	Error    string `json:"error" bson:"error"`
	Duration int64  `json:"duration" bson:"duration"`
}

type ExecutionLog struct {
	ID          bson.ObjectID  `json:"id" bson:"_id,omitempty"`
	ExecutionID bson.ObjectID  `json:"execution_id" bson:"execution_id"`
	NodeID      string         `json:"node_id" bson:"node_id"`
	NodeName    string         `json:"node_name" bson:"node_name"`
	NodeType    string         `json:"node_type" bson:"node_type"`
	Status      string         `json:"status" bson:"status"`
	Input       map[string]any `json:"input" bson:"input"`
	Output      map[string]any `json:"output" bson:"output"`
	Error       string         `json:"error" bson:"error"`
	Duration    int64          `json:"duration" bson:"duration"`
	AgentSteps  []AgentStep    `json:"agent_steps,omitempty" bson:"agent_steps,omitempty"`
	CreatedAt   time.Time      `json:"created_at" bson:"created_at"`
}

type AgentStep struct {
	Iteration  int            `json:"iteration" bson:"iteration"`
	Type       string         `json:"type" bson:"type"`
	Content    string         `json:"content,omitempty" bson:"content,omitempty"`
	ToolName   string         `json:"tool_name,omitempty" bson:"tool_name,omitempty"`
	ToolArgs   map[string]any `json:"tool_args,omitempty" bson:"tool_args,omitempty"`
	ToolResult string         `json:"tool_result,omitempty" bson:"tool_result,omitempty"`
}
