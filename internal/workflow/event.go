package workflow

import (
	"log"
	"time"
)

type EventType string

const (
	EventNodeStarted   EventType = "node_started"
	EventNodeCompleted EventType = "node_completed"
	EventNodeFailed    EventType = "node_failed"
	EventFlowCompleted EventType = "flow_completed"
	EventFlowFailed    EventType = "flow_failed"
)

// 执行事件
type Event struct {
	Type      EventType      `json:"type"`
	NodeID    string         `json:"node_id,omitempty"`
	NodeName  string         `json:"node_name,omitempty"`
	NodeType  NodeType       `json:"node_type,omitempty"`
	Output    map[string]any `json:"output,omitempty"`
	Error     string         `json:"error,omitempty"`
	Duration  int64          `json:"duration,omitempty"` // ms
	Timestamp time.Time      `json:"timestamp"`
}

// 事件总线，引擎往里写，SSE handler 从里读
type EventBus struct {
	ch chan Event
}

func NewEventBus() *EventBus {
	return &EventBus{ch: make(chan Event, 64)}
}

func (b *EventBus) Emit(e Event) {
	e.Timestamp = time.Now()
	select {
	case b.ch <- e:
	default:
		// channel 满了就丢弃，避免阻塞引擎
		log.Fatalf("event bus full")
	}
}

func (b *EventBus) Events() <-chan Event {
	return b.ch
}

func (b *EventBus) Close() {
	close(b.ch)
}
