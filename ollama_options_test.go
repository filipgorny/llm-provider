package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOllamaOptionsInRequest(t *testing.T) {
	var body map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_ = json.NewEncoder(w).Encode(ollamaGenerateResponse{Response: "ok"})
	}))

	defer srv.Close()

	o := NewOllama(srv.URL, "llama3")
	o.Options = map[string]any{"num_ctx": 8192}

	if _, err := o.Prompt(context.Background(), "hi"); err != nil {
		t.Fatalf("Prompt: %v", err)
	}

	opts, ok := body["options"].(map[string]any)

	if !ok {
		t.Fatalf("request missing options: %v", body)
	}

	if opts["num_ctx"] != float64(8192) {
		t.Errorf("num_ctx = %v, want 8192", opts["num_ctx"])
	}
}

func TestOllamaNoOptionsOmitted(t *testing.T) {
	var body map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_ = json.NewEncoder(w).Encode(ollamaGenerateResponse{Response: "ok"})
	}))

	defer srv.Close()

	o := NewOllama(srv.URL, "llama3")

	if _, err := o.Prompt(context.Background(), "hi"); err != nil {
		t.Fatalf("Prompt: %v", err)
	}

	if _, present := body["options"]; present {
		t.Errorf("options should be omitted when nil, got %v", body["options"])
	}
}
