package parse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OpenAIProvider struct {
	apiKey string
	model  string
}

func NewOpenAIProvider(apiKey, model string) *OpenAIProvider {
	if model == "" {
		model = "gpt-5.2"
	}
	return &OpenAIProvider{apiKey: apiKey, model: model}
}

func (p *OpenAIProvider) Name() string         { return "OpenAI" }
func (p *OpenAIProvider) ID() ProviderID       { return ProviderOpenAI }
func (p *OpenAIProvider) RequiresAPIKey() bool  { return true }
func (p *OpenAIProvider) DefaultEnvVar() string { return "OPENAI_API_KEY" }
func (p *OpenAIProvider) DefaultModel() string  { return "gpt-5.2" }

func (p *OpenAIProvider) Parse(query string) (string, string, error) {
	if p.apiKey == "" {
		return "", "", fmt.Errorf("OpenAI API key not set (set OPENAI_API_KEY or configure in settings with Ctrl+O)")
	}

	body := map[string]interface{}{
		"model": p.model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": fmt.Sprintf("Parse this song query and extract the artist and title. If any information is missing or misspelled, use your knowledge to complete it correctly. Query: %s", query),
			},
		},
		"response_format": map[string]interface{}{
			"type": "json_schema",
			"json_schema": map[string]interface{}{
				"name":   "song",
				"strict": true,
				"schema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"artist": map[string]string{"type": "string"},
						"title":  map[string]string{"type": "string"},
					},
					"required":             []string{"artist", "title"},
					"additionalProperties": false,
				},
			},
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("openai request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("openai returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", "", fmt.Errorf("failed to parse openai response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", "", fmt.Errorf("openai returned no choices")
	}

	var song parsedSong
	if err := json.Unmarshal([]byte(result.Choices[0].Message.Content), &song); err != nil {
		return "", "", fmt.Errorf("failed to parse song json: %w", err)
	}

	return song.Artist, song.Title, nil
}

func (p *OpenAIProvider) FetchLyrics(query string) (string, string, string, error) {
	if p.apiKey == "" {
		return "", "", "", fmt.Errorf("OpenAI API key not set (set OPENAI_API_KEY or configure in settings with Ctrl+O)")
	}

	body := map[string]interface{}{
		"model": p.model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": fmt.Sprintf("Identify the song and give me the full lyrics for: %s\nKeep empty lines between verses/chorus sections.", query),
			},
		},
		"response_format": map[string]interface{}{
			"type": "json_schema",
			"json_schema": map[string]interface{}{
				"name":   "lyrics",
				"strict": true,
				"schema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"artist": map[string]string{"type": "string"},
						"song":   map[string]string{"type": "string"},
						"lyrics": map[string]string{"type": "string"},
					},
					"required":             []string{"artist", "song", "lyrics"},
					"additionalProperties": false,
				},
			},
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", "", fmt.Errorf("openai request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", "", "", fmt.Errorf("openai returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", "", "", fmt.Errorf("failed to parse openai response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", "", "", fmt.Errorf("openai returned no choices")
	}

	var lr lyricsResult
	if err := json.Unmarshal([]byte(result.Choices[0].Message.Content), &lr); err != nil {
		return "", "", "", fmt.Errorf("failed to parse lyrics json: %w", err)
	}

	return lr.Artist, lr.Song, lr.Lyrics, nil
}
