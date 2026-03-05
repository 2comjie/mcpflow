package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

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
	ID          bson.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string        `json:"name" bson:"name"`
	Description string        `json:"description" bson:"description"`
	Nodes       []Node        `json:"nodes" bson:"nodes"`
	Edges       []Edge        `json:"edges" bson:"edges"`
	CreatedAt   time.Time     `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at" bson:"updated_at"`
}

// ==================== Node ====================

type Node struct {
	ID       string     `json:"id" bson:"id"`
	Type     NodeType   `json:"type" bson:"type"`
	Name     string     `json:"name" bson:"name"`
	Config   NodeConfig `json:"config" bson:"config"`
	Position Position   `json:"position" bson:"position"`
	Timeout  int        `json:"timeout,omitempty" bson:"timeout,omitempty"`
	Retry    *Retry     `json:"retry,omitempty" bson:"retry,omitempty"`
}

type Retry struct {
	Max      int `json:"max" bson:"max"`
	Interval int `json:"interval" bson:"interval"`
}

type Edge struct {
	ID        string `json:"id" bson:"id"`
	Source    string `json:"source" bson:"source"`
	Target    string `json:"target" bson:"target"`
	Condition string `json:"condition,omitempty" bson:"condition,omitempty"`
}

type Position struct {
	X float64 `json:"x" bson:"x"`
	Y float64 `json:"y" bson:"y"`
}

// ==================== NodeConfig ====================

type NodeConfig struct {
	LLM       *LLMConfig       `json:"llm,omitempty" bson:"llm,omitempty"`
	Agent     *AgentConfig     `json:"agent,omitempty" bson:"agent,omitempty"`
	Condition *ConditionConfig `json:"condition,omitempty" bson:"condition,omitempty"`
	Code      *CodeConfig      `json:"code,omitempty" bson:"code,omitempty"`
	HTTP      *HTTPConfig      `json:"http,omitempty" bson:"http,omitempty"`
	Email     *EmailConfig     `json:"email,omitempty" bson:"email,omitempty"`
}

type LLMConfig struct {
	BaseURL     string  `json:"base_url" bson:"base_url"`
	APIKey      string  `json:"api_key" bson:"api_key"`
	Model       string  `json:"model" bson:"model"`
	Prompt      string  `json:"prompt" bson:"prompt"`
	SystemMsg   string  `json:"system_msg,omitempty" bson:"system_msg,omitempty"`
	Temperature float64 `json:"temperature,omitempty" bson:"temperature,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty" bson:"max_tokens,omitempty"`
}

type AgentConfig struct {
	BaseURL       string           `json:"base_url" bson:"base_url"`
	APIKey        string           `json:"api_key" bson:"api_key"`
	Model         string           `json:"model" bson:"model"`
	Prompt        string           `json:"prompt" bson:"prompt"`
	SystemMsg     string           `json:"system_msg,omitempty" bson:"system_msg,omitempty"`
	McpServers    []AgentMCPServer `json:"mcp_servers" bson:"mcp_servers"`
	MaxIterations int              `json:"max_iterations,omitempty" bson:"max_iterations,omitempty"`
	Temperature   float64          `json:"temperature,omitempty" bson:"temperature,omitempty"`
	MaxTokens     int              `json:"max_tokens,omitempty" bson:"max_tokens,omitempty"`
}

type AgentMCPServer struct {
	URL     string            `json:"url" bson:"url"`
	Headers map[string]string `json:"headers,omitempty" bson:"headers,omitempty"`
}

type ConditionConfig struct {
	Expression string `json:"expression" bson:"expression"`
}

type CodeConfig struct {
	Language string `json:"language" bson:"language"`
	Code     string `json:"code" bson:"code"`
}

type HTTPConfig struct {
	Method  string            `json:"method" bson:"method"`
	URL     string            `json:"url" bson:"url"`
	Headers map[string]string `json:"headers,omitempty" bson:"headers,omitempty"`
	Body    string            `json:"body,omitempty" bson:"body,omitempty"`
}

type EmailConfig struct {
	SMTPHost    string `json:"smtp_host" bson:"smtp_host"`
	SMTPPort    int    `json:"smtp_port" bson:"smtp_port"`
	Username    string `json:"username" bson:"username"`
	Password    string `json:"password" bson:"password"`
	From        string `json:"from" bson:"from"`
	To          string `json:"to" bson:"to"`
	Cc          string `json:"cc,omitempty" bson:"cc,omitempty"`
	Subject     string `json:"subject" bson:"subject"`
	Body        string `json:"body" bson:"body"`
	ContentType string `json:"content_type,omitempty" bson:"content_type,omitempty"`
}
