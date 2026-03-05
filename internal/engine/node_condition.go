package engine

import (
	"fmt"

	"github.com/2comjie/mcpflow/internal/model"
	"github.com/expr-lang/expr"
)

func executeCondition(cfg *model.ConditionConfig, ctx *WorkflowContext) (any, error) {
	if cfg == nil {
		return nil, fmt.Errorf("condition config is nil")
	}

	env := map[string]any{
		"input": ctx.Input,
		"nodes": ctx.NodeOutput,
	}

	program, err := expr.Compile(cfg.Expression, expr.Env(env))
	if err != nil {
		return nil, fmt.Errorf("compile condition: %w", err)
	}

	result, err := expr.Run(program, env)
	if err != nil {
		return nil, fmt.Errorf("eval condition: %w", err)
	}

	b, ok := result.(bool)
	if !ok {
		return nil, fmt.Errorf("condition expression must return bool, got %T", result)
	}
	return b, nil
}
