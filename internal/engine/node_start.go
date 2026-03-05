package engine

func executeStart(ctx *WorkflowContext) (any, error) {
	return ctx.Input, nil
}
