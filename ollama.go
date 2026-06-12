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

// Ollama talks to an Ollama server over its HTTP API.
type Ollama struct {
	// URL is the base URL of the Ollama server, e.g. "http://localhost:11434".
	URL string

	// Model is the name of the model to use, e.g. "llama3".
	Model string

	// Client is the HTTP client used for requests. If nil, http.DefaultClient is used.
	Client *http.Client
}

var _ Llm = (*Ollama)(nil)

// NewOllama returns an Ollama backend for the given server URL and model.
func NewOllama(url, model string) *Ollama {
	return &Ollama{
		URL:    url,
		Model:  model,
		Client: http.DefaultClient,
	}
}

type ollamaGenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaGenerateResponse struct {
	Response string `json:"response"`
	Error    string `json:"error"`
}

// Prompt sends prompt to the Ollama server and returns the generated text.
func (o *Ollama) Prompt(ctx context.Context, prompt string) (string, error) {
	payload, err := json.Marshal(ollamaGenerateRequest{
		Model:  o.Model,
		Prompt: prompt,
		Stream: false,
	})

	if err != nil {
		return "", fmt.Errorf("ollama: marshal request: %w", err)
	}

	url := strings.TrimRight(o.URL, "/") + "/api/generate"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))

	if err != nil {
		return "", fmt.Errorf("ollama: build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := o.Client

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

	var out ollamaGenerateResponse

	if err := json.Unmarshal(data, &out); err != nil {
		return "", fmt.Errorf("ollama: decode response: %w", err)
	}

	if out.Error != "" {
		return "", fmt.Errorf("ollama: %s", out.Error)
	}

	return out.Response, nil
}
