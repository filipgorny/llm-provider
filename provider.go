package llm

import "context"

// LlmProvider is the Strategy context: it encapsulates a concrete Llm strategy
// and delegates calls to it. The strategy can be swapped at runtime, and the
// provider itself satisfies Llm so providers can be nested or decorated.
type LlmProvider struct {
	strategy Llm
}

var _ Llm = (*LlmProvider)(nil)

// NewLlmProvider returns a provider backed by the given strategy.
func NewLlmProvider(strategy Llm) *LlmProvider {
	return &LlmProvider{strategy: strategy}
}

// Prompt delegates to the underlying strategy.
func (p *LlmProvider) Prompt(ctx context.Context, prompt string) (string, error) {
	return p.strategy.Prompt(ctx, prompt)
}

// SetStrategy swaps the underlying strategy in place.
func (p *LlmProvider) SetStrategy(strategy Llm) {
	p.strategy = strategy
}

// Strategy returns the underlying strategy (introspection / testing).
func (p *LlmProvider) Strategy() Llm {
	return p.strategy
}
