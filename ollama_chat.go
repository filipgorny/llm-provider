package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var _ Chatter = (*Ollama)(nil)

// ollamaSession is a stateful conversation that resends the full history to the
// stateless Ollama /api/chat endpoint on every turn.
type ollamaSession struct {
	ollama   *Ollama
	messages []Message
}

type ollamaChatRequest struct {
	Model    string         `json:"model"`
	Messages []Message      `json:"messages"`
	Stream   bool           `json:"stream"`
	Options  map[string]any `json:"options,omitempty"`
}

type ollamaChatResponse struct {
	Message Message `json:"message"`
	Error   string  `json:"error"`
}

// NewSession starts an Ollama conversation. History is kept client-side and
// injected into every request.
func (o *Ollama) NewSession(opts ...SessionOption) Session {
	cfg := newSessionConfig(opts)

	var messages []Message

	if cfg.system != "" {
		messages = append(messages, Message{Role: "system", Content: cfg.system})
	}

	messages = append(messages, cfg.history...)

	return &ollamaSession{ollama: o, messages: messages}
}

func (s *ollamaSession) History() []Message {
	return s.messages
}

func (s *ollamaSession) Send(ctx context.Context, text string) (string, error) {
	s.messages = append(s.messages, Message{Role: "user", Content: text})

	payload, err := json.Marshal(ollamaChatRequest{
		Model:    s.ollama.Model,
		Messages: s.messages,
		Stream:   false,
		Options:  s.ollama.Options,
	})

	if err != nil {
		return "", fmt.Errorf("ollama: marshal request: %w", err)
	}

	url := strings.TrimRight(s.ollama.URL, "/") + "/api/chat"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))

	if err != nil {
		return "", fmt.Errorf("ollama: build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := s.ollama.Client

	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)

	if err != nil {
		return "", fmt.Errorf("ollama: do request: %w", err)
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", fmt.Errorf("ollama: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama: unexpected status %d: %s", resp.StatusCode, string(data))
	}

	var out ollamaChatResponse

	if err := json.Unmarshal(data, &out); err != nil {
		return "", fmt.Errorf("ollama: decode response: %w", err)
	}

	if out.Error != "" {
		return "", fmt.Errorf("ollama: %s", out.Error)
	}

	s.messages = append(s.messages, out.Message)

	return out.Message.Content, nil
}
