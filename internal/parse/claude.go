package parse

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

// ClaudeParser uses Claude AI to parse song queries into structured artist/title.
type ClaudeParser struct{}

// NewClaudeParser creates a new Claude-based song parser.
func NewClaudeParser() *ClaudeParser {
	return &ClaudeParser{}
}

type parsedSong struct {
	Artist string `json:"artist"`
	Title  string `json:"title"`
}

type claudeResponse struct {
	StructuredOutput parsedSong `json:"structured_output"`
}

// Parse extracts artist and title from a natural language query using Claude.
func (p *ClaudeParser) Parse(query string) (artist, title string, err error) {
	prompt := fmt.Sprintf("Parse this song query and extract the artist and title. If any information is missing or misspelled, use your knowledge to complete it correctly: %s", query)

	cmd := exec.Command("claude", "-p", prompt,
		"--output-format", "json",
		"--json-schema", `{"type":"object","properties":{"artist":{"type":"string"},"title":{"type":"string"}},"required":["artist","title"]}`)

	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("claude command failed: %w", err)
	}

	var response claudeResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return "", "", fmt.Errorf("failed to parse claude response: %w", err)
	}

	return response.StructuredOutput.Artist, response.StructuredOutput.Title, nil
}
