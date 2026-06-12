package llm

// ToolCall is a tool invocation the model decided to make.
type ToolCall struct {
	Name      string
	Arguments map[string]any
}
