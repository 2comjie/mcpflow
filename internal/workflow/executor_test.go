package workflow

import (
	"context"
	"testing"
)

func TestConditionExecutor_Equal(t *testing.T) {
	executor := &ConditionExecutor{}
	node := &Node{
		Type: NodeCondition,
		Config: NodeConfig{
			Condition: &ConditionConfig{
				Expression: `status == "ok"`,
			},
		},
	}

	input := map[string]any{"status": "ok"}
	output, err := executor.Execute(context.Background(), node, input)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if output["branch"] != "true" {
		t.Fatalf("expected branch true, got %v", output["branch"])
	}
}

func TestConditionExecutor_NotEqual(t *testing.T) {
	executor := &ConditionExecutor{}
	node := &Node{
		Type: NodeCondition,
		Config: NodeConfig{
			Condition: &ConditionConfig{
				Expression: `status == "ok"`,
			},
		},
	}

	input := map[string]any{"status": "fail"}
	output, err := executor.Execute(context.Background(), node, input)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if output["branch"] != "false" {
		t.Fatalf("expected branch false, got %v", output["branch"])
	}
}

func TestConditionExecutor_NumericCompare(t *testing.T) {
	executor := &ConditionExecutor{}
	node := &Node{
		Type: NodeCondition,
		Config: NodeConfig{
			Condition: &ConditionConfig{
				Expression: `count > 5`,
			},
		},
	}

	input := map[string]any{"count": 10}
	output, err := executor.Execute(context.Background(), node, input)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if output["branch"] != "true" {
		t.Fatalf("expected branch true, got %v", output["branch"])
	}
}

func TestConditionExecutor_ComplexExpr(t *testing.T) {
	executor := &ConditionExecutor{}
	node := &Node{
		Type: NodeCondition,
		Config: NodeConfig{
			Condition: &ConditionConfig{
				Expression: `count > 5 && status == "ok"`,
			},
		},
	}

	input := map[string]any{"count": 10, "status": "ok"}
	output, err := executor.Execute(context.Background(), node, input)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if output["branch"] != "true" {
		t.Fatalf("expected branch true, got %v", output["branch"])
	}
}

func TestConditionExecutor_NestedField(t *testing.T) {
	executor := &ConditionExecutor{}
	node := &Node{
		Type: NodeCondition,
		Config: NodeConfig{
			Condition: &ConditionConfig{
				Expression: `user.age >= 18`,
			},
		},
	}

	input := map[string]any{
		"user": map[string]any{"age": 20},
	}
	output, err := executor.Execute(context.Background(), node, input)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if output["branch"] != "true" {
		t.Fatalf("expected branch true, got %v", output["branch"])
	}
}

func TestConditionExecutor_NilConfig(t *testing.T) {
	executor := &ConditionExecutor{}
	node := &Node{Type: NodeCondition, Config: NodeConfig{}}

	_, err := executor.Execute(context.Background(), node, nil)
	if err == nil {
		t.Fatal("expected error for nil config")
	}
}

// ==================== Engine 集成测试 ====================

func TestEngine_SimpleFlow(t *testing.T) {
	// start -> end
	registry := &ExecutorRegistry{executors: map[NodeType]NodeExecutor{
		NodeStart: &StartExecutor{},
		NodeEnd:   &EndExecutor{},
	}}
	engine := NewEngine(registry)

	wf := &Workflow{
		Nodes: Nodes{
			{ID: "s", Type: NodeStart, Name: "开始"},
			{ID: "e", Type: NodeEnd, Name: "结束"},
		},
		Edges: Edges{
			{ID: "e1", Source: "s", Target: "e"},
		},
	}

	output, states, err := engine.Run(context.Background(), wf, map[string]any{"msg": "hello"}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if output["msg"] != "hello" {
		t.Fatalf("expected msg=hello, got %v", output["msg"])
	}
	if len(states) != 2 {
		t.Fatalf("expected 2 node states, got %d", len(states))
	}
}

func TestEngine_ConditionBranch(t *testing.T) {
	// start -> condition -> (true) end
	registry := &ExecutorRegistry{executors: map[NodeType]NodeExecutor{
		NodeStart:     &StartExecutor{},
		NodeEnd:       &EndExecutor{},
		NodeCondition: &ConditionExecutor{},
	}}
	engine := NewEngine(registry)

	wf := &Workflow{
		Nodes: Nodes{
			{ID: "s", Type: NodeStart, Name: "开始"},
			{ID: "c", Type: NodeCondition, Name: "判断", Config: NodeConfig{
				Condition: &ConditionConfig{Expression: `score > 60`},
			}},
			{ID: "pass", Type: NodeEnd, Name: "通过"},
			{ID: "fail", Type: NodeEnd, Name: "不通过"},
		},
		Edges: Edges{
			{ID: "e1", Source: "s", Target: "c"},
			{ID: "e2", Source: "c", Target: "pass", Condition: "true"},
			{ID: "e3", Source: "c", Target: "fail", Condition: "false"},
		},
	}

	// score=80 走 true 分支
	output, states, err := engine.Run(context.Background(), wf, map[string]any{"score": 80}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if _, ok := states["pass"]; !ok {
		t.Fatal("expected pass node to be executed")
	}
	if _, ok := states["fail"]; ok {
		t.Fatal("expected fail node NOT to be executed")
	}
	_ = output
}

// ==================== CodeExecutor 测试 ====================

func TestCodeExecutor_ReturnMap(t *testing.T) {
	executor := &CodeExecutor{}
	node := &Node{
		Type: NodeCode,
		Config: NodeConfig{
			Code: &CodeConfig{
				Language: "javascript",
				Code:     `({result: input.a + input.b})`,
			},
		},
	}

	input := map[string]any{"a": 3, "b": 5}
	output, err := executor.Execute(context.Background(), node, input)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	// goja 返回 int64
	if output["result"] != int64(8) {
		t.Fatalf("expected result=8, got %v (%T)", output["result"], output["result"])
	}
}

func TestCodeExecutor_ReturnScalar(t *testing.T) {
	executor := &CodeExecutor{}
	node := &Node{
		Type: NodeCode,
		Config: NodeConfig{
			Code: &CodeConfig{
				Language: "javascript",
				Code:     `input.name + " world"`,
			},
		},
	}

	input := map[string]any{"name": "hello"}
	output, err := executor.Execute(context.Background(), node, input)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if output["result"] != "hello world" {
		t.Fatalf("expected 'hello world', got %v", output["result"])
	}
}

func TestCodeExecutor_SyntaxError(t *testing.T) {
	executor := &CodeExecutor{}
	node := &Node{
		Type: NodeCode,
		Config: NodeConfig{
			Code: &CodeConfig{
				Language: "javascript",
				Code:     `var x = !!!;`,
			},
		},
	}

	_, err := executor.Execute(context.Background(), node, map[string]any{})
	if err == nil {
		t.Fatal("expected error for syntax error")
	}
}

func TestCodeExecutor_Lua_ReturnTable(t *testing.T) {
	executor := &CodeExecutor{}
	node := &Node{
		Type: NodeCode,
		Config: NodeConfig{
			Code: &CodeConfig{
				Language: "lua",
				Code:     `output = {result = input.a + input.b, doubled = true}`,
			},
		},
	}

	input := map[string]any{"a": 3, "b": 5}
	output, err := executor.Execute(context.Background(), node, input)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if output["result"] != float64(8) {
		t.Fatalf("expected result=8, got %v (%T)", output["result"], output["result"])
	}
	if output["doubled"] != true {
		t.Fatalf("expected doubled=true, got %v", output["doubled"])
	}
}

func TestCodeExecutor_Lua_ReturnScalar(t *testing.T) {
	executor := &CodeExecutor{}
	node := &Node{
		Type: NodeCode,
		Config: NodeConfig{
			Code: &CodeConfig{
				Language: "lua",
				Code:     `output = input.name .. " world"`,
			},
		},
	}

	input := map[string]any{"name": "hello"}
	output, err := executor.Execute(context.Background(), node, input)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if output["result"] != "hello world" {
		t.Fatalf("expected 'hello world', got %v", output["result"])
	}
}

func TestCodeExecutor_Lua_SyntaxError(t *testing.T) {
	executor := &CodeExecutor{}
	node := &Node{
		Type: NodeCode,
		Config: NodeConfig{
			Code: &CodeConfig{
				Language: "lua",
				Code:     `output = {{{}`,
			},
		},
	}

	_, err := executor.Execute(context.Background(), node, map[string]any{})
	if err == nil {
		t.Fatal("expected error for lua syntax error")
	}
}

func TestCodeExecutor_UnsupportedLang(t *testing.T) {
	executor := &CodeExecutor{}
	node := &Node{
		Type: NodeCode,
		Config: NodeConfig{
			Code: &CodeConfig{Language: "ruby", Code: "puts 1"},
		},
	}

	_, err := executor.Execute(context.Background(), node, map[string]any{})
	if err == nil {
		t.Fatal("expected error for unsupported language")
	}
}

func TestCodeExecutor_NilConfig(t *testing.T) {
	executor := &CodeExecutor{}
	node := &Node{Type: NodeCode, Config: NodeConfig{}}

	_, err := executor.Execute(context.Background(), node, nil)
	if err == nil {
		t.Fatal("expected error for nil config")
	}
}

func TestEngine_NoStartNode(t *testing.T) {
	registry := &ExecutorRegistry{executors: map[NodeType]NodeExecutor{}}
	engine := NewEngine(registry)

	wf := &Workflow{
		Nodes: Nodes{{ID: "e", Type: NodeEnd, Name: "结束"}},
	}
	_, _, err := engine.Run(context.Background(), wf, nil, nil, nil)
	if err == nil {
		t.Fatal("expected error for no start node")
	}
}
