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

var _ ToolCaller = (*Ollama)(nil)

type ollamaTool struct {
	Type     string             `json:"type"`
	Function ollamaToolFunction `json:"function"`
}

type ollamaToolFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

type ollamaToolChatRequest struct {
	Model    string         `json:"model"`
	Messages []Message      `json:"messages"`
	Tools    []ollamaTool   `json:"tools,omitempty"`
	Stream   bool           `json:"stream"`
	Options  map[string]any `json:"options,omitempty"`
}

type ollamaRespToolCall struct {
	Function struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	} `json:"function"`
}

type ollamaToolChatResponse struct {
	Message struct {
		Role      string               `json:"role"`
		Content   string               `json:"content"`
		ToolCalls []ollamaRespToolCall `json:"tool_calls"`
	} `json:"message"`
	Error string `json:"error"`
}

// CallTools performs a native tool-calling chat against the Ollama server.
func (o *Ollama) CallTools(ctx context.Context, messages []Message, tools []ToolSpec) (ToolResponse, error) {
	specs := make([]ollamaTool, 0, len(tools))

	for _, t := range tools {
		specs = append(specs, ollamaTool{
			Type: "function",
			Function: ollamaToolFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			},
		})
	}

	payload, err := json.Marshal(ollamaToolChatRequest{
		Model:    o.Model,
		Messages: messages,
		Tools:    specs,
		Stream:   false,
		Options:  o.Options,
	})

	if err != nil {
		return ToolResponse{}, fmt.Errorf("ollama: marshal tool request: %w", err)
	}

	url := strings.TrimRight(o.URL, "/") + "/api/chat"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))

	if err != nil {
		return ToolResponse{}, fmt.Errorf("ollama: build tool request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := o.Client

	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)

	if err != nil {
		return ToolResponse{}, fmt.Errorf("ollama: do tool request: %w", err)
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)

	if err != nil {
		return ToolResponse{}, fmt.Errorf("ollama: read tool response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if strings.Contains(string(data), "does not support tools") {
			return ToolResponse{}, fmt.Errorf("%w: %s", ErrToolsUnsupported, strings.TrimSpace(string(data)))
		}

		return ToolResponse{}, fmt.Errorf("ollama: unexpected status %d: %s", resp.StatusCode, string(data))
	}

	var out ollamaToolChatResponse

	if err := json.Unmarshal(data, &out); err != nil {
		return ToolResponse{}, fmt.Errorf("ollama: decode tool response: %w", err)
	}

	if out.Error != "" {
		return ToolResponse{}, fmt.Errorf("ollama: %s", out.Error)
	}

	result := ToolResponse{Text: out.Message.Content}

	for _, c := range out.Message.ToolCalls {
		result.Calls = append(result.Calls, ToolCall{
			Name:      c.Function.Name,
			Arguments: c.Function.Arguments,
		})
	}

	return result, nil
}
