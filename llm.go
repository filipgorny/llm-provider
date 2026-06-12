// Package llm defines a minimal interface for large language model backends
// together with implementations for different providers.
package llm

import "context"

// Llm is the strategy interface for a text-in, text-out language model backend.
// Concrete strategies (Ollama, ClaudeHeadless) are injected into an LlmProvider.
type Llm interface {
	// Prompt sends a single prompt and returns the model's response.
	Prompt(ctx context.Context, prompt string) (string, error)
}
