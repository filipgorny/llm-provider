//go:build integration

package llm

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestClaudeHeadlessPromptIntegration runs the real claude CLI in headless
// mode. It is skipped when the claude binary is not on PATH. Optionally set
// CLAUDE_MODEL to pin a specific model.
//
// Run with: go test -tags integration ./...
func TestClaudeHeadlessPromptIntegration(t *testing.T) {
	if _, err := exec.LookPath("claude"); err != nil {
		t.Skip("claude binary not found on PATH")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)

	defer cancel()

	c := NewClaudeHeadless()

	got, err := c.Prompt(ctx, "Reply with a single word: ping")

	if err != nil {
		t.Fatalf("Prompt: %v", err)
	}

	if got == "" {
		t.Error("expected a non-empty response")
	}

	t.Logf("claude response: %q", got)
}

func TestClaudeChatIntegration(t *testing.T) {
	if _, err := exec.LookPath("claude"); err != nil {
		t.Skip("claude binary not found on PATH")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 240*time.Second)

	defer cancel()

	sess := NewClaudeHeadless().NewSession()

	r1, err := sess.Send(ctx, "Zapamiętaj liczbę 4291. Odpowiedz tylko: ok")

	if err != nil {
		t.Fatalf("turn 1: %v", err)
	}

	t.Logf("turn 1: %q", r1)

	// Turn 2 asks about the number WITHOUT repeating it — only --resume carries it.
	r2, err := sess.Send(ctx, "Jaką liczbę kazałem Ci zapamiętać? Odpowiedz samą liczbą.")

	if err != nil {
		t.Fatalf("turn 2: %v", err)
	}

	t.Logf("turn 2 (recalled via resume): %q", r2)

	if !strings.Contains(r2, "4291") {
		t.Errorf("expected the recalled number 4291 in %q", r2)
	}
}
