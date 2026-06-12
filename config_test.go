package llm

import "testing"

func TestConfigOllama(t *testing.T) {
	p, err := NewLlmProviderFromConfig(Config{
		Llm:    "ollama",
		Ollama: OllamaConfig{URL: "http://localhost:11434", Model: "llama3"},
	})

	if err != nil {
		t.Fatalf("build: %v", err)
	}

	o, ok := p.Strategy().(*Ollama)

	if !ok {
		t.Fatalf("strategy = %T, want *Ollama", p.Strategy())
	}

	if o.URL != "http://localhost:11434" {
		t.Errorf("url = %q", o.URL)
	}

	if o.Model != "llama3" {
		t.Errorf("model = %q", o.Model)
	}
}

func TestConfigClaudeWithModel(t *testing.T) {
	p, err := NewLlmProviderFromConfig(Config{
		Llm:    "claude",
		Claude: ClaudeConfig{Model: "claude-opus-4-8"},
	})

	if err != nil {
		t.Fatalf("build: %v", err)
	}

	c, ok := p.Strategy().(*ClaudeHeadless)

	if !ok {
		t.Fatalf("strategy = %T, want *ClaudeHeadless", p.Strategy())
	}

	if c.Model != "claude-opus-4-8" {
		t.Errorf("model = %q, want claude-opus-4-8", c.Model)
	}
}

func TestConfigClaudeNoModel(t *testing.T) {
	p, err := NewLlmProviderFromConfig(Config{Llm: "claude"})

	if err != nil {
		t.Fatalf("build: %v", err)
	}

	c := p.Strategy().(*ClaudeHeadless)

	if c.Model != "" {
		t.Errorf("model = %q, want empty (CLI default)", c.Model)
	}
}

func TestConfigErrors(t *testing.T) {
	cases := map[string]Config{
		"unknown llm":   {Llm: "gpt"},
		"no llm":        {},
		"ollama no url": {Llm: "ollama", Ollama: OllamaConfig{Model: "llama3"}},
		"ollama no model": {
			Llm:    "ollama",
			Ollama: OllamaConfig{URL: "http://localhost:11434"},
		},
	}

	for name, c := range cases {

		t.Run(name, func(t *testing.T) {
			_, err := NewLlmProviderFromConfig(c)

			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestYamlBytesSource(t *testing.T) {
	yaml := []byte("llm: ollama\nollama:\n  url: http://localhost:11434\n  model: qwen2.5-coder:14b\n")

	c, err := YamlBytes(yaml).Load()

	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if c.Llm != "ollama" {
		t.Errorf("llm = %q", c.Llm)
	}

	if c.Ollama.URL != "http://localhost:11434" {
		t.Errorf("url = %q", c.Ollama.URL)
	}

	if c.Ollama.Model != "qwen2.5-coder:14b" {
		t.Errorf("model = %q", c.Ollama.Model)
	}
}

func TestEnvSource(t *testing.T) {
	t.Setenv("LLM", "ollama")
	t.Setenv("LLM_OLLAMA_URL", "http://localhost:11434")
	t.Setenv("LLM_OLLAMA_MODEL", "llama3")

	p, err := NewLlmProviderFrom(Env("LLM"))

	if err != nil {
		t.Fatalf("build: %v", err)
	}

	o, ok := p.Strategy().(*Ollama)

	if !ok {
		t.Fatalf("strategy = %T, want *Ollama", p.Strategy())
	}

	if o.URL != "http://localhost:11434" || o.Model != "llama3" {
		t.Errorf("ollama config = %+v", o)
	}
}
