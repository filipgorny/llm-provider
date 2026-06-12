package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOllamaPrompt(t *testing.T) {
	var gotReq ollamaGenerateRequest

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" {
			t.Errorf("path = %q, want /api/generate", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)

		if err := json.Unmarshal(body, &gotReq); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		_ = json.NewEncoder(w).Encode(ollamaGenerateResponse{Response: "pong"})
	}))

	defer srv.Close()

	o := NewOllama(srv.URL, "llama3")

	got, err := o.Prompt(context.Background(), "ping")

	if err != nil {
		t.Fatalf("Prompt: %v", err)
	}

	if got != "pong" {
		t.Errorf("response = %q, want %q", got, "pong")
	}

	if gotReq.Model != "llama3" {
		t.Errorf("model = %q, want llama3", gotReq.Model)
	}

	if gotReq.Prompt != "ping" {
		t.Errorf("prompt = %q, want ping", gotReq.Prompt)
	}

	if gotReq.Stream {
		t.Error("stream = true, want false")
	}
}

func TestOllamaPromptHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, "boom")
	}))

	defer srv.Close()

	o := NewOllama(srv.URL, "llama3")

	_, err := o.Prompt(context.Background(), "ping")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestOllamaPromptAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(ollamaGenerateResponse{Error: "model not found"})
	}))

	defer srv.Close()

	o := NewOllama(srv.URL, "ghost")

	_, err := o.Prompt(context.Background(), "ping")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
