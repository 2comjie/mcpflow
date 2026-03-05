package engine

func executeEnd(ctx *WorkflowContext) (any, error) {
	return getPreviousOutput(ctx), nil
}
