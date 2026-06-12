package llm

import (
	"context"
	"testing"
)

// fakeLlm is a recording strategy used to verify provider delegation.
type fakeLlm struct {
	gotPrompt string
	reply     string
}

func (f *fakeLlm) Prompt(ctx context.Context, prompt string) (string, error) {
	f.gotPrompt = prompt

	return f.reply, nil
}

func TestLlmProviderDelegates(t *testing.T) {
	strat := &fakeLlm{reply: "pong"}

	p := NewLlmProvider(strat)

	got, err := p.Prompt(context.Background(), "ping")

	if err != nil {
		t.Fatalf("Prompt: %v", err)
	}

	if got != "pong" {
		t.Errorf("response = %q, want pong", got)
	}

	if strat.gotPrompt != "ping" {
		t.Errorf("strategy got prompt %q, want ping", strat.gotPrompt)
	}
}

func TestLlmProviderSetStrategy(t *testing.T) {
	first := &fakeLlm{reply: "first"}
	second := &fakeLlm{reply: "second"}

	p := NewLlmProvider(first)

	p.SetStrategy(second)

	if p.Strategy() != second {
		t.Error("Strategy() did not return the swapped strategy")
	}

	got, err := p.Prompt(context.Background(), "x")

	if err != nil {
		t.Fatalf("Prompt: %v", err)
	}

	if got != "second" {
		t.Errorf("response = %q, want second", got)
	}
}
