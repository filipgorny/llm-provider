package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOllamaSessionInjectsHistory(t *testing.T) {
	var requests []ollamaChatRequest

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Errorf("path = %q, want /api/chat", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)

		var req ollamaChatRequest

		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		requests = append(requests, req)

		reply := Message{Role: "assistant", Content: "Warszawa"}

		if len(requests) == 2 {
			reply.Content = "ok. około 1.86 mln"
		}

		_ = json.NewEncoder(w).Encode(ollamaChatResponse{Message: reply})
	}))

	defer srv.Close()

	provider := NewLlmProvider(NewOllama(srv.URL, "llama3"))

	sess, err := provider.NewSession()

	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	if _, err := sess.Send(context.Background(), "Jaka jest stolica Polski?"); err != nil {
		t.Fatalf("turn 1: %v", err)
	}

	if _, err := sess.Send(context.Background(), "A ile ma mieszkańców?"); err != nil {
		t.Fatalf("turn 2: %v", err)
	}

	// Turn 2 must resend the full prior history (Ollama is stateless).
	second := requests[1].Messages

	if len(second) != 3 {
		t.Fatalf("turn 2 messages = %d, want 3 (user, assistant, user)", len(second))
	}

	if second[0].Content != "Jaka jest stolica Polski?" || second[0].Role != "user" {
		t.Errorf("turn 2 msg[0] = %+v", second[0])
	}

	if second[1].Role != "assistant" || second[1].Content != "Warszawa" {
		t.Errorf("turn 2 msg[1] = %+v", second[1])
	}

	// Local history holds all four turns.
	if h := sess.History(); len(h) != 4 {
		t.Errorf("history = %d, want 4", len(h))
	}
}

func TestOllamaSessionSystemPrompt(t *testing.T) {
	var gotReq ollamaChatRequest

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotReq)
		_ = json.NewEncoder(w).Encode(ollamaChatResponse{Message: Message{Role: "assistant", Content: "ok"}})
	}))

	defer srv.Close()

	provider := NewLlmProvider(NewOllama(srv.URL, "llama3"))

	sess, _ := provider.NewSession(WithSystemPrompt("Odpowiadaj po polsku."))

	if _, err := sess.Send(context.Background(), "hej"); err != nil {
		t.Fatalf("Send: %v", err)
	}

	if len(gotReq.Messages) < 1 || gotReq.Messages[0].Role != "system" {
		t.Fatalf("first message = %+v, want role=system", gotReq.Messages)
	}

	if gotReq.Messages[0].Content != "Odpowiadaj po polsku." {
		t.Errorf("system content = %q", gotReq.Messages[0].Content)
	}
}
