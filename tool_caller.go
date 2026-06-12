package llm

import "context"

// ToolCaller is the optional capability of a strategy to do native tool calling:
// the model receives tool schemas and returns structured tool calls. Backends
// that don't implement it (e.g. ClaudeHeadless) signal "no native tool calling".
type ToolCaller interface {
	CallTools(ctx context.Context, messages []Message, tools []ToolSpec) (ToolResponse, error)
}
