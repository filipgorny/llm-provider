package llm

import "os"

// envSource loads a Config from environment variables under a prefix. For a
// prefix of "LLM" it reads:
//
//	LLM               -> Config.Llm        ("ollama" | "claude")
//	LLM_OLLAMA_URL    -> Config.Ollama.URL
//	LLM_OLLAMA_MODEL  -> Config.Ollama.Model
//	LLM_CLAUDE_MODEL  -> Config.Claude.Model
type envSource struct {
	prefix string
}

// Env returns a ConfigSource backed by environment variables under prefix.
func Env(prefix string) ConfigSource {
	return envSource{prefix: prefix}
}

func (s envSource) Load() (Config, error) {
	p := s.prefix

	return Config{
		Llm: os.Getenv(p),
		Ollama: OllamaConfig{
			URL:   os.Getenv(p + "_OLLAMA_URL"),
			Model: os.Getenv(p + "_OLLAMA_MODEL"),
		},
		Claude: ClaudeConfig{
			Model: os.Getenv(p + "_CLAUDE_MODEL"),
		},
	}, nil
}
