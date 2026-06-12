package llm

// ToolResponse is the model's reply to a tool-calling request: either tool calls
// to execute, or final assistant text (Calls empty).
type ToolResponse struct {
	Text  string
	Calls []ToolCall
}
