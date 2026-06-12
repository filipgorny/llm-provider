package llm

// ToolSpec describes a tool exposed to the model for native tool calling.
// Parameters is a JSON Schema object describing the tool's arguments.
type ToolSpec struct {
	Name        string
	Description string
	Parameters  map[string]any
}
