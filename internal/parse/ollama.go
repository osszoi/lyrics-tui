package parse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OllamaProvider struct {
	model string
}

func NewOllamaProvider(model string) *OllamaProvider {
	if model == "" {
		model = "qwen2.5-coder:14b"
	}
	return &OllamaProvider{model: model}
}

func (p *OllamaProvider) Name() string         { return "Ollama" }
func (p *OllamaProvider) ID() ProviderID       { return ProviderOllama }
func (p *OllamaProvider) RequiresAPIKey() bool  { return false }
func (p *OllamaProvider) DefaultEnvVar() string { return "" }
func (p *OllamaProvider) DefaultModel() string  { return "qwen2.5-coder:14b" }

func (p *OllamaProvider) Parse(query string) (string, string, error) {
	prompt := fmt.Sprintf("Parse this song query and extract the artist and title. If any information is missing or misspelled, use your knowledge to complete it correctly. Respond with JSON only: {\"artist\": \"...\", \"title\": \"...\"}. Query: %s", query)

	body := map[string]interface{}{
		"model":  p.model,
		"prompt": prompt,
		"format": "json",
		"stream": false,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "http://localhost:11434/api/generate", bytes.NewReader(jsonBody))
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("ollama request failed (is ollama running?): %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Response string `json:"response"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", "", fmt.Errorf("failed to parse ollama response: %w", err)
	}

	var song parsedSong
	if err := json.Unmarshal([]byte(result.Response), &song); err != nil {
		return "", "", fmt.Errorf("failed to parse song json: %w", err)
	}

	return song.Artist, song.Title, nil
}

func (p *OllamaProvider) FetchLyrics(query string) (string, string, string, error) {
	body := map[string]interface{}{
		"model":  p.model,
		"prompt": fmt.Sprintf("Identify the song and give me the full lyrics for: %s\nKeep empty lines between verses/chorus sections. Respond with JSON only: {\"artist\": \"...\", \"song\": \"...\", \"lyrics\": \"...\"}", query),
		"format": "json",
		"stream": false,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "http://localhost:11434/api/generate", bytes.NewReader(jsonBody))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", "", fmt.Errorf("ollama request failed (is ollama running?): %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", "", "", fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Response string `json:"response"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", "", "", fmt.Errorf("failed to parse ollama response: %w", err)
	}

	var lr lyricsResult
	if err := json.Unmarshal([]byte(result.Response), &lr); err != nil {
		return "", "", "", fmt.Errorf("failed to parse lyrics json: %w", err)
	}

	return lr.Artist, lr.Song, lr.Lyrics, nil
}
