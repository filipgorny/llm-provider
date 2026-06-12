package llm

import "fmt"

// Config selects which LLM strategy to use and carries per-backend parameters.
// It can be built programmatically, loaded from YAML, or read from the
// environment — see ConfigSource and its implementations.
type Config struct {
	// Llm selects the strategy: "ollama" or "claude".
	Llm string `yaml:"llm"`

	// Ollama holds parameters used when Llm == "ollama".
	Ollama OllamaConfig `yaml:"ollama"`

	// Claude holds parameters used when Llm == "claude".
	Claude ClaudeConfig `yaml:"claude"`
}

// OllamaConfig configures the Ollama strategy. URL and Model are required.
type OllamaConfig struct {
	URL   string `yaml:"url"`
	Model string `yaml:"model"`

	// Options are passed through as Ollama request options, e.g.
	// {num_ctx: 8192} to size the context window, or {temperature: 0.2}.
	Options map[string]any `yaml:"options"`
}

// ClaudeConfig configures the ClaudeHeadless strategy. Model is optional; when
// empty, the claude CLI default model is used.
type ClaudeConfig struct {
	Model string `yaml:"model"`
}

// ConfigSource is anything that can produce a Config. YAML is the default
// source, but env and programmatic Config values are equally valid.
type ConfigSource interface {
	Load() (Config, error)
}

// NewLlmProviderFromConfig builds a provider from an in-memory Config. This is
// the main entry point when importing the library and passing config from code.
func NewLlmProviderFromConfig(c Config) (*LlmProvider, error) {
	switch c.Llm {

	case "ollama":

		if c.Ollama.URL == "" {
			return nil, fmt.Errorf("llm: ollama requires a url")
		}

		if c.Ollama.Model == "" {
			return nil, fmt.Errorf("llm: ollama requires a model")
		}

		ollama := NewOllama(c.Ollama.URL, c.Ollama.Model)
		ollama.Options = c.Ollama.Options

		return NewLlmProvider(ollama), nil

	case "claude":

		strategy := NewClaudeHeadless()

		if c.Claude.Model != "" {
			strategy.Model = c.Claude.Model
		}

		return NewLlmProvider(strategy), nil

	case "":
		return nil, fmt.Errorf("llm: no llm selected (set the \"llm\" field)")

	default:
		return nil, fmt.Errorf("llm: unknown llm %q", c.Llm)
	}
}

// NewLlmProviderFrom loads a Config from any source and builds a provider.
// For the default YAML source: NewLlmProviderFrom(YamlFile(path)).
func NewLlmProviderFrom(src ConfigSource) (*LlmProvider, error) {
	c, err := src.Load()

	if err != nil {
		return nil, err
	}

	return NewLlmProviderFromConfig(c)
}
