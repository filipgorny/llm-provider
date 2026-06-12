//go:build integration

package llm

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestOllamaPromptIntegration hits a real Ollama server. Configure it with:
//
//	OLLAMA_URL   base URL (default "http://localhost:11434")
//	OLLAMA_MODEL model name (default "llama3")
//
// Run with: go test -tags integration ./...
func ollamaIntegrationTarget() (url, model string) {
	url = os.Getenv("OLLAMA_URL")

	if url == "" {
		url = "http://localhost:11434"
	}

	model = os.Getenv("OLLAMA_MODEL")

	if model == "" {
		model = "llama3"
	}

	return url, model
}

func TestOllamaPromptIntegration(t *testing.T) {
	url, model := ollamaIntegrationTarget()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

	defer cancel()

	o := NewOllama(url, model)

	got, err := o.Prompt(ctx, "Reply with a single word: ping")

	if err != nil {
		t.Fatalf("Prompt: %v", err)
	}

	if got == "" {
		t.Error("expected a non-empty response")
	}

	t.Logf("ollama response: %q", got)
}

func TestOllamaChatIntegration(t *testing.T) {
	url, model := ollamaIntegrationTarget()

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)

	defer cancel()

	sess := NewOllama(url, model).NewSession()

	r1, err := sess.Send(ctx, "Jaka jest stolica Polski? Odpowiedz jednym słowem.")

	if err != nil {
		t.Fatalf("turn 1: %v", err)
	}

	t.Logf("turn 1: %q", r1)

	r2, err := sess.Send(ctx, "A w jakim kraju leży to miasto? Odpowiedz jednym słowem.")

	if err != nil {
		t.Fatalf("turn 2: %v", err)
	}

	t.Logf("turn 2 (follow-up using context): %q", r2)

	if r2 == "" {
		t.Error("expected a non-empty follow-up response")
	}
}
