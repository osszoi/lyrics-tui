package parse

import (
	"fmt"
	"os"

	"lyrics-tui/internal/config"
)

type ProviderID string

const (
	ProviderOpenAI ProviderID = "openai"
	ProviderGemini ProviderID = "gemini"
	ProviderOllama ProviderID = "ollama"
)

var AllProviders = []ProviderID{ProviderOpenAI, ProviderGemini, ProviderOllama}

type Provider interface {
	Name() string
	ID() ProviderID
	Parse(query string) (artist, title string, err error)
	FetchLyrics(query string) (artist, title, lyrics string, err error)
	RequiresAPIKey() bool
	DefaultEnvVar() string
	DefaultModel() string
}

func ProviderName(id ProviderID) string {
	switch id {
	case ProviderOpenAI:
		return "OpenAI"
	case ProviderGemini:
		return "Gemini"
	case ProviderOllama:
		return "Ollama"
	default:
		return string(id)
	}
}

func DefaultModelForProvider(id ProviderID) string {
	switch id {
	case ProviderOpenAI:
		return "gpt-5.2"
	case ProviderGemini:
		return "gemini-3-pro-preview"
	case ProviderOllama:
		return "qwen2.5-coder:14b"
	default:
		return ""
	}
}

func NewProviderFromConfig(cfg *config.Config) (Provider, error) {
	id := ProviderID(cfg.Provider)
	model := cfg.Model
	if model == "" {
		model = DefaultModelForProvider(id)
	}

	switch id {
	case ProviderOpenAI:
		key := cfg.APIKey
		if key == "" {
			key = os.Getenv("OPENAI_API_KEY")
		}
		return NewOpenAIProvider(key, model), nil
	case ProviderGemini:
		key := cfg.APIKey
		if key == "" {
			key = os.Getenv("GEMINI_API_KEY")
		}
		return NewGeminiProvider(key, model), nil
	case ProviderOllama:
		return NewOllamaProvider(model), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", cfg.Provider)
	}
}
