package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/2comjie/mcpflow/pkg/types"
)

type JSON = types.JSONMap

type NodeType string

const (
	NodeStart     NodeType = "start"
	NodeEnd       NodeType = "end"
	NodeCondition NodeType = "condition"
	NodeCode      NodeType = "code"
	NodeLLM       NodeType = "llm"
	NodeAgent     NodeType = "agent"
	NodeHTTP      NodeType = "http"
	NodeEmail     NodeType = "email"
)

// ==================== Workflow ====================

type Workflow struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:255;not null"`
	Description string    `json:"description" gorm:"type:text"`
	Nodes       Nodes     `json:"nodes" gorm:"type:json"`
	Edges       Edges     `json:"edges" gorm:"type:json"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Workflow) TableName() string { return "workflows" }

// ==================== Node ====================

type Node struct {
	ID       string     `json:"id"`
	Type     NodeType   `json:"type"`
	Name     string     `json:"name"`
	Config   NodeConfig `json:"config"`
	Position Position   `json:"position"`
	Timeout  int        `json:"timeout,omitempty"`
	Retry    *Retry     `json:"retry,omitempty"`
}

type Retry struct {
	Max      int `json:"max"`
	Interval int `json:"interval"`
}

type Edge struct {
	ID        string `json:"id"`
	Source    string `json:"source"`
	Target    string `json:"target"`
	Condition string `json:"condition,omitempty"`
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// ==================== NodeConfig ====================

type NodeConfig struct {
	LLM       *LLMConfig       `json:"llm,omitempty"`
	Agent     *AgentConfig     `json:"agent,omitempty"`
	Condition *ConditionConfig `json:"condition,omitempty"`
	Code      *CodeConfig      `json:"code,omitempty"`
	HTTP      *HTTPConfig      `json:"http,omitempty"`
	Email     *EmailConfig     `json:"email,omitempty"`
}

type LLMConfig struct {
	BaseURL     string  `json:"base_url"`
	APIKey      string  `json:"api_key"`
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	SystemMsg   string  `json:"system_msg,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
}

type AgentConfig struct {
	BaseURL       string           `json:"base_url"`
	APIKey        string           `json:"api_key"`
	Model         string           `json:"model"`
	Prompt        string           `json:"prompt"`
	SystemMsg     string           `json:"system_msg,omitempty"`
	McpServers    []AgentMCPServer `json:"mcp_servers"`
	MaxIterations int              `json:"max_iterations,omitempty"`
	Temperature   float64          `json:"temperature,omitempty"`
	MaxTokens     int              `json:"max_tokens,omitempty"`
}

type AgentMCPServer struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}

type ConditionConfig struct {
	Expression string `json:"expression"`
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

type EmailConfig struct {
	SMTPHost    string `json:"smtp_host"`
	SMTPPort    int    `json:"smtp_port"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	From        string `json:"from"`
	To          string `json:"to"`
	Cc          string `json:"cc,omitempty"`
	Subject     string `json:"subject"`
	Body        string `json:"body"`
	ContentType string `json:"content_type,omitempty"`
}

// ==================== GORM JSON 自定义类型 ====================

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
