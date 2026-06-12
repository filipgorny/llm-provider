package llm

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestClaudeHeadlessPrompt(t *testing.T) {
	var (
		gotName string
		gotArgs []string
	)

	c := &ClaudeHeadless{
		Bin:   "claude",
		Model: "claude-opus-4-8",
		runner: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			gotName = name
			gotArgs = args

			return []byte("  hello from claude\n"), nil
		},
	}

	got, err := c.Prompt(context.Background(), "say hi")

	if err != nil {
		t.Fatalf("Prompt: %v", err)
	}

	if got != "hello from claude" {
		t.Errorf("response = %q, want %q", got, "hello from claude")
	}

	if gotName != "claude" {
		t.Errorf("bin = %q, want claude", gotName)
	}

	want := []string{"-p", "say hi", "--model", "claude-opus-4-8"}

	if !reflect.DeepEqual(gotArgs, want) {
		t.Errorf("args = %v, want %v", gotArgs, want)
	}
}

func TestClaudeHeadlessPromptNoModel(t *testing.T) {
	var gotArgs []string

	c := &ClaudeHeadless{
		runner: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			gotArgs = args

			return []byte("ok"), nil
		},
	}

	_, err := c.Prompt(context.Background(), "hi")

	if err != nil {
		t.Fatalf("Prompt: %v", err)
	}

	want := []string{"-p", "hi"}

	if !reflect.DeepEqual(gotArgs, want) {
		t.Errorf("args = %v, want %v", gotArgs, want)
	}
}

func TestClaudeHeadlessPromptError(t *testing.T) {
	c := &ClaudeHeadless{
		runner: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			return nil, errors.New("boom")
		},
	}

	_, err := c.Prompt(context.Background(), "x")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
