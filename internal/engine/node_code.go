package engine

import (
	"encoding/json"
	"fmt"

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

	// 注入上下文变量
	_ = vm.Set("input", ctx.Input)
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

	val, err := vm.RunString(code)
	if err != nil {
		return nil, fmt.Errorf("js execution error: %w", err)
	}

	return val.Export(), nil
}
