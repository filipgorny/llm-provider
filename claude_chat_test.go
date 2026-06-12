package llm

import (
	"context"
	"slices"
	"testing"
)

func TestClaudeSessionResumesWithoutInjectingHistory(t *testing.T) {
	var calls [][]string

	c := &ClaudeHeadless{
		Bin: "claude",
		runner: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			calls = append(calls, args)

			if len(calls) == 1 {
				return []byte(`{"result":"Warszawa","session_id":"sid-123"}`), nil
			}

			return []byte(`{"result":"około 1.86 mln","session_id":"sid-123"}`), nil
		},
	}

	sess, err := NewLlmProvider(c).NewSession()

	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	r1, err := sess.Send(context.Background(), "Jaka jest stolica Polski?")

	if err != nil {
		t.Fatalf("turn 1: %v", err)
	}

	if r1 != "Warszawa" {
		t.Errorf("turn 1 = %q, want Warszawa", r1)
	}

	r2, err := sess.Send(context.Background(), "A ile ma mieszkańców?")

	if err != nil {
		t.Fatalf("turn 2: %v", err)
	}

	if r2 != "około 1.86 mln" {
		t.Errorf("turn 2 = %q", r2)
	}

	// Turn 1: no --resume yet.
	if slices.Contains(calls[0], "--resume") {
		t.Error("turn 1 should not contain --resume")
	}

	// Turn 2: resume by the captured session id...
	if !slices.Contains(calls[1], "--resume") || !slices.Contains(calls[1], "sid-123") {
		t.Errorf("turn 2 args missing --resume sid-123: %v", calls[1])
	}

	// ...and must NOT re-inject the prior turn's text (Claude keeps it server-side).
	if slices.Contains(calls[1], "Jaka jest stolica Polski?") {
		t.Errorf("turn 2 must not resend turn 1 text: %v", calls[1])
	}

	if h := sess.History(); len(h) != 4 {
		t.Errorf("history = %d, want 4", len(h))
	}
}

func TestClaudeSessionSystemPrompt(t *testing.T) {
	var gotArgs []string

	c := &ClaudeHeadless{
		runner: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			gotArgs = args

			return []byte(`{"result":"ok","session_id":"sid"}`), nil
		},
	}

	sess, _ := NewLlmProvider(c).NewSession(WithSystemPrompt("Bądź zwięzły."))

	if _, err := sess.Send(context.Background(), "hej"); err != nil {
		t.Fatalf("Send: %v", err)
	}

	if !slices.Contains(gotArgs, "--append-system-prompt") {
		t.Errorf("args missing --append-system-prompt: %v", gotArgs)
	}
}
