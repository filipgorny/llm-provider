package llm

import "testing"

func TestProviderNewSessionUnsupported(t *testing.T) {
	// fakeLlm (from provider_test.go) implements only Prompt, not Chatter.
	p := NewLlmProvider(&fakeLlm{})

	_, err := p.NewSession()

	if err == nil {
		t.Fatal("expected error for non-chat strategy, got nil")
	}
}
