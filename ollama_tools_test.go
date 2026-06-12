package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOllamaCallTools(t *testing.T) {
	var body map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)

		_, _ = w.Write([]byte(`{"message":{"role":"assistant","content":"","tool_calls":[{"function":{"name":"file_read","arguments":{"path":"x"}}}]}}`))
	}))

	defer srv.Close()

	o := NewOllama(srv.URL, "qwen3")

	resp, err := o.CallTools(context.Background(),
		[]Message{{Role: "user", Content: "read x"}},
		[]ToolSpec{{Name: "file_read", Description: "read a file", Parameters: map[string]any{"type": "object"}}})

	if err != nil {
		t.Fatalf("CallTools: %v", err)
	}

	if len(resp.Calls) != 1 || resp.Calls[0].Name != "file_read" {
		t.Fatalf("calls = %+v", resp.Calls)
	}

	if resp.Calls[0].Arguments["path"] != "x" {
		t.Errorf("args = %v", resp.Calls[0].Arguments)
	}

	// tools must be present in the request body.
	if _, ok := body["tools"]; !ok {
		t.Errorf("request missing tools: %v", body)
	}
}

func TestOllamaCallToolsFinalText(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"message":{"role":"assistant","content":"the answer is 42"}}`))
	}))

	defer srv.Close()

	o := NewOllama(srv.URL, "qwen3")

	resp, err := o.CallTools(context.Background(), []Message{{Role: "user", Content: "q"}}, nil)

	if err != nil {
		t.Fatalf("CallTools: %v", err)
	}

	if len(resp.Calls) != 0 || resp.Text != "the answer is 42" {
		t.Errorf("resp = %+v", resp)
	}
}

var _ ToolCaller = (*Ollama)(nil)
