package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Provider string
	APIKey   string
	Model    string
}

func DefaultConfig() *Config {
	return &Config{
		Provider: "ollama",
		Model:    "qwen2.5-coder:14b",
	}
}

func configDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "lyrics")
}

func configPath() string {
	return filepath.Join(configDir(), "config.toml")
}

func Load() *Config {
	cfg := DefaultConfig()
	data, err := os.ReadFile(configPath())
	if err != nil {
		return cfg
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"")
		switch key {
		case "provider":
			cfg.Provider = value
		case "api_key":
			cfg.APIKey = value
		case "model":
			cfg.Model = value
		}
	}
	return cfg
}

func (c *Config) Save() error {
	dir := configDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	content := fmt.Sprintf("provider = \"%s\"\napi_key = \"%s\"\nmodel = \"%s\"\n",
		c.Provider, c.APIKey, c.Model)
	return os.WriteFile(configPath(), []byte(content), 0644)
}
