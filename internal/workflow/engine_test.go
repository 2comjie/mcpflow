package workflow

import (
	"context"
	"fmt"
	"testing"
)

// 构造一个最简工作流: start -> end
func TestEngine_SimpleFlow(t *testing.T) {
	registry := &ExecutorRegistry{executors: make(map[NodeType]NodeExecutor)}
	registry.Register(NodeStart, &StartExecutor{})
	registry.Register(NodeEnd, &EndExecutor{})
	engine := NewEngine(registry)

	wf := &Workflow{
		Nodes: Nodes{
			{ID: "n1", Type: NodeStart, Name: "开始"},
			{ID: "n2", Type: NodeEnd, Name: "结束"},
		},
		Edges: Edges{
			{ID: "e1", Source: "n1", Target: "n2"},
		},
	}

	input := map[string]any{"msg": "hello"}
	output, states, err := engine.Run(context.Background(), wf, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// end 节点透传输入，所以 output 应该等于 input
	if output["msg"] != "hello" {
		t.Errorf("expected output msg=hello, got %v", output["msg"])
	}

	// 两个节点都应该是 completed
	if states["n1"].Status != string(ExecCompleted) {
		t.Errorf("n1 status expected completed, got %s", states["n1"].Status)
	}
	if states["n2"].Status != string(ExecCompleted) {
		t.Errorf("n2 status expected completed, got %s", states["n2"].Status)
	}
}

// 三节点链路: start -> mock节点 -> end
func TestEngine_ThreeNodeChain(t *testing.T) {
	registry := &ExecutorRegistry{executors: make(map[NodeType]NodeExecutor)}
	registry.Register(NodeStart, &StartExecutor{})
	registry.Register(NodeEnd, &EndExecutor{})
	registry.Register(NodeMCPTool, &mockExecutor{
		output: map[string]any{"tool_result": "search done"},
	})
	engine := NewEngine(registry)

	wf := &Workflow{
		Nodes: Nodes{
			{ID: "n1", Type: NodeStart, Name: "开始"},
			{ID: "n2", Type: NodeMCPTool, Name: "工具调用", Config: NodeConfig{
				MCPTool: &MCPToolConfig{ServerURL: "http://fake", ToolName: "search"},
			}},
			{ID: "n3", Type: NodeEnd, Name: "结束"},
		},
		Edges: Edges{
			{ID: "e1", Source: "n1", Target: "n2"},
			{ID: "e2", Source: "n2", Target: "n3"},
		},
	}

	output, states, err := engine.Run(context.Background(), wf, map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output["tool_result"] != "search done" {
		t.Errorf("expected tool_result='search done', got %v", output["tool_result"])
	}

	// 三个节点都完成
	for _, id := range []string{"n1", "n2", "n3"} {
		if states[id].Status != string(ExecCompleted) {
			t.Errorf("node %s expected completed, got %s", id, states[id].Status)
		}
	}
}

// 条件分支: start -> condition -> (true: nodeA) -> end
func TestEngine_ConditionBranch(t *testing.T) {
	registry := &ExecutorRegistry{executors: make(map[NodeType]NodeExecutor)}
	registry.Register(NodeStart, &StartExecutor{})
	registry.Register(NodeEnd, &EndExecutor{})
	registry.Register(NodeCondition, &mockExecutor{
		output: map[string]any{"branch": "yes"},
	})
	registry.Register(NodeMCPTool, &mockExecutor{
		output: map[string]any{"result": "yes branch"},
	})
	registry.Register(NodeHTTP, &mockExecutor{
		output: map[string]any{"result": "no branch"},
	})
	engine := NewEngine(registry)

	wf := &Workflow{
		Nodes: Nodes{
			{ID: "n1", Type: NodeStart, Name: "开始"},
			{ID: "n2", Type: NodeCondition, Name: "判断", Config: NodeConfig{
				Condition: &ConditionConfig{Expression: "true"},
			}},
			{ID: "n3", Type: NodeMCPTool, Name: "是分支", Config: NodeConfig{
				MCPTool: &MCPToolConfig{},
			}},
			{ID: "n4", Type: NodeHTTP, Name: "否分支", Config: NodeConfig{
				HTTP: &HTTPConfig{},
			}},
			{ID: "n5", Type: NodeEnd, Name: "结束"},
		},
		Edges: Edges{
			{ID: "e1", Source: "n1", Target: "n2"},
			{ID: "e2", Source: "n2", Target: "n3", Condition: "yes"},
			{ID: "e3", Source: "n2", Target: "n4", Condition: "no"},
			{ID: "e4", Source: "n3", Target: "n5"},
			{ID: "e5", Source: "n4", Target: "n5"},
		},
	}

	output, states, err := engine.Run(context.Background(), wf, map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 应该走 yes 分支 (n3)，不走 no 分支 (n4)
	if output["result"] != "yes branch" {
		t.Errorf("expected 'yes branch', got %v", output["result"])
	}
	if _, ran := states["n4"]; ran {
		t.Error("n4 (no branch) should not have been executed")
	}
}

// 模板变量测试
func TestEngine_TemplateVariables(t *testing.T) {
	registry := &ExecutorRegistry{executors: make(map[NodeType]NodeExecutor)}
	registry.Register(NodeStart, &StartExecutor{})
	registry.Register(NodeEnd, &EndExecutor{})
	// n2 输出固定结果
	registry.Register(NodeHTTP, &mockExecutor{
		output: map[string]any{"body": "天气晴朗"},
	})
	// n3 捕获渲染后的 input 来验证模板是否生效
	captureExec := &captureExecutor{}
	registry.Register(NodeLLM, captureExec)
	engine := NewEngine(registry)

	wf := &Workflow{
		Nodes: Nodes{
			{ID: "n1", Type: NodeStart, Name: "开始"},
			{ID: "n2", Type: NodeHTTP, Name: "获取天气", Config: NodeConfig{
				HTTP: &HTTPConfig{Method: "GET", URL: "http://fake"},
			}},
			{ID: "n3", Type: NodeLLM, Name: "总结", Config: NodeConfig{
				LLM: &LLMConfig{Model: "gpt-4", Prompt: "总结: {{n2.body}}"},
			}},
			{ID: "n4", Type: NodeEnd, Name: "结束"},
		},
		Edges: Edges{
			{ID: "e1", Source: "n1", Target: "n2"},
			{ID: "e2", Source: "n2", Target: "n3"},
			{ID: "e3", Source: "n3", Target: "n4"},
		},
	}

	_, _, err := engine.Run(context.Background(), wf, map[string]any{"query": "今天天气"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 验证 n3 收到的 input 中 "body" 字段已经被模板渲染成 n2 的输出
	got, _ := captureExec.lastInput["body"].(string)
	if got != "天气晴朗" {
		t.Errorf("expected template rendered to '天气晴朗', got '%s'", got)
	}
}

// 无 start 节点应该报错
func TestEngine_NoStartNode(t *testing.T) {
	engine := NewEngine(&ExecutorRegistry{executors: make(map[NodeType]NodeExecutor)})
	wf := &Workflow{
		Nodes: Nodes{{ID: "n1", Type: NodeEnd, Name: "结束"}},
	}
	_, _, err := engine.Run(context.Background(), wf, nil)
	if err == nil {
		t.Fatal("expected error for missing start node")
	}
}

// 节点执行失败应该传播错误
func TestEngine_NodeFailure(t *testing.T) {
	registry := &ExecutorRegistry{executors: make(map[NodeType]NodeExecutor)}
	registry.Register(NodeStart, &StartExecutor{})
	registry.Register(NodeEnd, &EndExecutor{})
	registry.Register(NodeCode, &failExecutor{})
	engine := NewEngine(registry)

	wf := &Workflow{
		Nodes: Nodes{
			{ID: "n1", Type: NodeStart, Name: "开始"},
			{ID: "n2", Type: NodeCode, Name: "出错节点", Config: NodeConfig{
				Code: &CodeConfig{Language: "js", Code: "fail"},
			}},
			{ID: "n3", Type: NodeEnd, Name: "结束"},
		},
		Edges: Edges{
			{ID: "e1", Source: "n1", Target: "n2"},
			{ID: "e2", Source: "n2", Target: "n3"},
		},
	}

	_, states, err := engine.Run(context.Background(), wf, map[string]any{})
	if err == nil {
		t.Fatal("expected error from failing node")
	}
	if states["n2"].Status != string(ExecFailed) {
		t.Errorf("n2 expected failed, got %s", states["n2"].Status)
	}
	// n3 不应该被执行
	if _, ran := states["n3"]; ran {
		t.Error("n3 should not have been executed after n2 failure")
	}
}

// ==================== 测试辅助 ====================

// mockExecutor 返回固定输出
type mockExecutor struct {
	output map[string]any
}

func (m *mockExecutor) Execute(ctx context.Context, node *Node, input map[string]any) (map[string]any, error) {
	return m.output, nil
}

// captureExecutor 捕获输入，返回固定输出
type captureExecutor struct {
	lastInput map[string]any
}

func (c *captureExecutor) Execute(ctx context.Context, node *Node, input map[string]any) (map[string]any, error) {
	c.lastInput = input
	return map[string]any{"summary": "done"}, nil
}

// failExecutor 总是失败
type failExecutor struct{}

func (f *failExecutor) Execute(ctx context.Context, node *Node, input map[string]any) (map[string]any, error) {
	return nil, fmt.Errorf("execution failed")
}
