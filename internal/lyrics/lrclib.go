package lyrics

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// LRCLIBProvider fetches synced lyrics from lrclib.net.
type LRCLIBProvider struct {
	client *http.Client
}

// NewLRCLIBProvider creates a new LRCLIB lyrics provider.
func NewLRCLIBProvider() *LRCLIBProvider {
	return &LRCLIBProvider{
		client: &http.Client{},
	}
}

type lrclibResponse struct {
	SyncedLyrics string `json:"syncedLyrics"`
}

// FetchLyrics is not supported by LRCLIB (only synced lyrics).
func (p *LRCLIBProvider) FetchLyrics(artist, title string) (string, error) {
	return "", fmt.Errorf("lrclib only provides synced lyrics")
}

// FetchSynced retrieves time-synced lyrics from LRCLIB.
func (p *LRCLIBProvider) FetchSynced(artist, title string) ([]Line, error) {
	apiURL := fmt.Sprintf("https://lrclib.net/api/get?artist_name=%s&track_name=%s",
		url.QueryEscape(artist),
		url.QueryEscape(title))

	resp, err := p.client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("lrclib request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("lrclib returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read lrclib response: %w", err)
	}

	var lrcResp lrclibResponse
	if err := json.Unmarshal(body, &lrcResp); err != nil {
		return nil, fmt.Errorf("failed to parse lrclib response: %w", err)
	}

	if lrcResp.SyncedLyrics == "" {
		return nil, fmt.Errorf("no synced lyrics available")
	}

	return ParseLRC(lrcResp.SyncedLyrics), nil
}
