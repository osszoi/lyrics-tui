package parse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type GeminiProvider struct {
	apiKey string
	model  string
}

func NewGeminiProvider(apiKey, model string) *GeminiProvider {
	if model == "" {
		model = "gemini-3-pro-preview"
	}
	return &GeminiProvider{apiKey: apiKey, model: model}
}

func (p *GeminiProvider) Name() string          { return "Gemini" }
func (p *GeminiProvider) ID() ProviderID        { return ProviderGemini }
func (p *GeminiProvider) RequiresAPIKey() bool  { return true }
func (p *GeminiProvider) DefaultEnvVar() string { return "GEMINI_API_KEY" }
func (p *GeminiProvider) DefaultModel() string  { return "gemini-3-pro-preview" }

func (p *GeminiProvider) Parse(query string) (string, string, error) {
	if p.apiKey == "" {
		return "", "", fmt.Errorf("Gemini API key not set (set GEMINI_API_KEY or configure in settings with Ctrl+O)")
	}

	body := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{
						"text": fmt.Sprintf("Parse this song query and extract the artist and title. If any information is missing or misspelled, use your knowledge to complete it correctly. Respond with JSON only: {\"artist\": \"...\", \"title\": \"...\"}. Query: %s", query),
					},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"responseMimeType": "application/json",
			"responseSchema": map[string]interface{}{
				"type": "OBJECT",
				"properties": map[string]interface{}{
					"artist": map[string]string{"type": "STRING"},
					"title":  map[string]string{"type": "STRING"},
				},
				"required": []string{"artist", "title"},
			},
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", p.model, p.apiKey)
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("gemini request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("gemini returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", "", fmt.Errorf("failed to parse gemini response: %w", err)
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", "", fmt.Errorf("gemini returned no content")
	}

	var song parsedSong
	if err := json.Unmarshal([]byte(result.Candidates[0].Content.Parts[0].Text), &song); err != nil {
		return "", "", fmt.Errorf("failed to parse song json: %w", err)
	}

	return song.Artist, song.Title, nil
}

func (p *GeminiProvider) FetchLyrics(query string) (string, string, string, error) {
	if p.apiKey == "" {
		return "", "", "", fmt.Errorf("Gemini API key not set (set GEMINI_API_KEY or configure in settings with Ctrl+O)")
	}

	body := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{
						"text": fmt.Sprintf("Identify the song and give me the full lyrics for: %s\nKeep empty lines between verses/chorus sections.", query),
					},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"responseMimeType": "application/json",
			"responseSchema": map[string]interface{}{
				"type": "OBJECT",
				"properties": map[string]interface{}{
					"artist": map[string]string{"type": "STRING"},
					"song":   map[string]string{"type": "STRING"},
					"lyrics": map[string]string{"type": "STRING"},
				},
				"required": []string{"artist", "song", "lyrics"},
			},
			"thinkingConfig": map[string]interface{}{
				"thinkingLevel": "MINIMAL",
			},
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", p.model, p.apiKey)
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", "", fmt.Errorf("gemini request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", "", "", fmt.Errorf("gemini returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", "", "", fmt.Errorf("failed to parse gemini response: %w", err)
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", "", "", fmt.Errorf("gemini returned no content")
	}

	var lr lyricsResult
	if err := json.Unmarshal([]byte(result.Candidates[0].Content.Parts[0].Text), &lr); err != nil {
		return "", "", "", fmt.Errorf("failed to parse lyrics json: %w", err)
	}

	return lr.Artist, lr.Song, lr.Lyrics, nil
}
