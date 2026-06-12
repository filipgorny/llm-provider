package llm

import (
	"context"
	"fmt"
)

// Message is one turn in a conversation.
type Message struct {
	Role    string `json:"role"` // "system" | "user" | "assistant"
	Content string `json:"content"`
}

// Session is a stateful, multi-turn conversation with a backend: each Send
// knows the context of previous turns in the same session.
type Session interface {
	// Send sends the next user turn and returns the assistant's reply.
	Send(ctx context.Context, text string) (string, error)

	// History returns the conversation so far.
	History() []Message
}

// Chatter is the optional capability of a strategy to start a Session. Not
// every Llm strategy must support chat.
type Chatter interface {
	NewSession(opts ...SessionOption) Session
}

// SessionOption configures a new Session.
type SessionOption func(*sessionConfig)

type sessionConfig struct {
	system  string
	history []Message
}

// WithSystemPrompt sets a system prompt for the conversation.
func WithSystemPrompt(s string) SessionOption {
	return func(c *sessionConfig) {
		c.system = s
	}
}

// WithHistory seeds the conversation with prior messages (e.g. to resume).
func WithHistory(h []Message) SessionOption {
	return func(c *sessionConfig) {
		c.history = h
	}
}

func newSessionConfig(opts []SessionOption) sessionConfig {
	var c sessionConfig

	for _, opt := range opts {
		opt(&c)
	}

	return c
}

// NewSession starts a conversation if the underlying strategy supports chat,
// otherwise it returns an error.
func (p *LlmProvider) NewSession(opts ...SessionOption) (Session, error) {
	c, ok := p.strategy.(Chatter)

	if !ok {
		return nil, fmt.Errorf("llm: strategy %T does not support chat", p.strategy)
	}

	return c.NewSession(opts...), nil
}
