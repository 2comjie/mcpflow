package engine

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/2comjie/mcpflow/internal/model"
	"github.com/dop251/goja"
)

func executeCode(cfg *model.CodeConfig, ctx *WorkflowContext) (any, error) {
	if cfg == nil {
		return nil, fmt.Errorf("code config is nil")
	}

	switch cfg.Language {
	case "javascript", "js", "":
		return executeJS(cfg.Code, ctx)
	default:
		return nil, fmt.Errorf("unsupported language: %s", cfg.Language)
	}
}

func executeJS(code string, ctx *WorkflowContext) (any, error) {
	vm := goja.New()

	// 执行超时保护（30 秒）
	timer := time.AfterFunc(30*time.Second, func() {
		vm.Interrupt("execution timeout (30s)")
	})
	defer timer.Stop()

	// 注入上下文变量
	// input = 上一个节点的输出（更直觉），workflow_input = 工作流原始输入
	prevOutput := getPreviousOutput(ctx)
	if prevOutput != nil {
		_ = vm.Set("input", prevOutput)
	} else {
		_ = vm.Set("input", ctx.Input)
	}
	_ = vm.Set("workflow_input", ctx.Input)
	_ = vm.Set("nodes", ctx.NodeOutput)

	// 注入 JSON 工具
	_ = vm.Set("JSON_stringify", func(v any) string {
		b, _ := json.Marshal(v)
		return string(b)
	})
	_ = vm.Set("JSON_parse", func(s string) any {
		var v any
		_ = json.Unmarshal([]byte(s), &v)
		return v
	})

	// 包装为函数调用，支持 return 语句
	wrapped := "(function() {\n" + code + "\n})()"
	val, err := vm.RunString(wrapped)
	if err != nil {
		return nil, fmt.Errorf("js execution error: %w", err)
	}

	return val.Export(), nil
}
