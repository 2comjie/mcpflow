package workflow

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/2comjie/mcpflow/pkg/types"
)

type NodeType string

const (
	NodeStart       NodeType = "start"
	NodeEnd         NodeType = "end"
	NodeMCPTool     NodeType = "mcp_tool"
	NodeMCPPrompt   NodeType = "mcp_prompt"
	NodeMCPResource NodeType = "mcp_resource"
	NodeLLM         NodeType = "llm"
	NodeCondition   NodeType = "condition"
	NodeLoop        NodeType = "loop"
	NodeParallel    NodeType = "parallel"
	NodeCode        NodeType = "code"
	NodeHTTP        NodeType = "http"
)

type ExecStatus string

const (
	ExecPending   ExecStatus = "pending"
	ExecRunning   ExecStatus = "running"
	ExecCompleted ExecStatus = "completed"
	ExecFailed    ExecStatus = "failed"
	ExecCancelled ExecStatus = "cancelled"
)

// JSON 通用 JSON 字段类型，引用公共包
type JSON = types.JSONMap

// Nodes 节点列表 JSON 类型
type Nodes []Node

func (n Nodes) Value() (driver.Value, error) {
	if n == nil {
		return "[]", nil
	}
	b, err := json.Marshal(n)
	return string(b), err
}

func (n *Nodes) Scan(value any) error {
	if value == nil {
		*n = make(Nodes, 0)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan Nodes: %v", value)
	}
	return json.Unmarshal(bytes, n)
}

// Edges 边列表 JSON 类型
type Edges []Edge

func (e Edges) Value() (driver.Value, error) {
	if e == nil {
		return "[]", nil
	}
	b, err := json.Marshal(e)
	return string(b), err
}

func (e *Edges) Scan(value any) error {
	if value == nil {
		*e = make(Edges, 0)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan Edges: %v", value)
	}
	return json.Unmarshal(bytes, e)
}

// NodeStates 节点状态映射 JSON 类型
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

// ==================== GORM 模型 ====================
type Workflow struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:255;not null"`
	Description string    `json:"description" gorm:"type:text"`
	Status      string    `json:"status" gorm:"size:20;default:draft"`
	Nodes       Nodes     `json:"nodes" gorm:"type:json"`
	Edges       Edges     `json:"edges" gorm:"type:json"`
	Variables   JSON      `json:"variables" gorm:"type:json"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (w *Workflow) TableName() string {
	return "workflows"
}

// 执行实例
type WorkflowExecution struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	WorkflowID uint       `json:"workflow_id" gorm:"index;not null"`
	Status     ExecStatus `json:"status" gorm:"size:20;default:pending"`
	Input      JSON       `json:"input" gorm:"type:json"`
	Output     JSON       `json:"output" gorm:"type:json"`
	Context    JSON       `json:"context" gorm:"type:json"`
	NodeStates NodeStates `json:"node_states" gorm:"type:json"`
	Error      string     `json:"error" gorm:"type:text"`
	StartedAt  *time.Time `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

// 节点
type Node struct {
	ID       string     `json:"id"`
	Type     NodeType   `json:"type"`
	Name     string     `json:"name"`
	Config   NodeConfig `json:"config"`
	Position Position   `json:"position"`
	Timeout  int        `json:"timeout,omitempty"` // 秒，0 表示用默认值
}

type Edge struct {
	ID         string `json:"id"`
	Source     string `json:"source"`
	Target     string `json:"target"`
	SourcePort string `json:"source_port,omitempty"`
	TargetPort string `json:"target_port,omitempty"`
	Condition  string `json:"condition,omitempty"`
}

// 画布位置
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// ==================== 节点配置 ====================
type NodeConfig struct {
	MCPTool     *MCPToolConfig     `json:"mcp_tool,omitempty"`
	MCPPrompt   *MCPPromptConfig   `json:"mcp_prompt,omitempty"`
	LLM         *LLMConfig         `json:"llm,omitempty"`
	Condition   *ConditionConfig   `json:"condition,omitempty"`
	Loop        *LoopConfig        `json:"loop,omitempty"`
	Code        *CodeConfig        `json:"code,omitempty"`
	HTTP        *HTTPConfig        `json:"http,omitempty"`
	MCPResource *MCPResourceConfig `json:"mcp_resource,omitempty"`
}

// MCP工具调用配置
type MCPToolConfig struct {
	ServerURL string         `json:"server_url"`
	ToolName  string         `json:"tool_name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

// MCP提示词配置
type MCPPromptConfig struct {
	ServerURL  string         `json:"server_url"`
	PromptName string         `json:"prompt_name"`
	Arguments  map[string]any `json:"arguments,omitempty"`
}

// LLM调用配置
type LLMConfig struct {
	Provider    string  `json:"provider"`
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	SystemMsg   string  `json:"system_msg,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
}

// MCP资源读取配置
type MCPResourceConfig struct {
	ServerURL string `json:"server_url"`
	URI       string `json:"uri"`
}

// 条件分支配置
type ConditionConfig struct {
	Expression string            `json:"expression"`
	Branches   map[string]string `json:"branches"` // "true"->"nodeId", "false"->"nodeId"
}

type LoopConfig struct {
	Condition string `json:"condition"`
	MaxCount  int    `json:"max_count"`
	ItemsFrom string `json:"items_from,omitempty"`
}

type CodeConfig struct {
	Language string `json:"language"`
	Code     string `json:"code"`
}

type HTTPConfig struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
}

// ==================== 执行状态 ====================

// NodeState 单个节点的执行状态
type NodeState struct {
	NodeID   string `json:"node_id"`
	Status   string `json:"status"`
	Input    any    `json:"input,omitempty"`
	Output   any    `json:"output,omitempty"`
	Error    string `json:"error,omitempty"`
	Duration int64  `json:"duration"` // 毫秒
}
