package lyrics

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// GeniusProvider fetches lyrics from Genius.com.
type GeniusProvider struct {
	accessToken string
	client      *http.Client
}

// NewGeniusProvider creates a new Genius lyrics provider.
func NewGeniusProvider(accessToken string) *GeniusProvider {
	return &GeniusProvider{
		accessToken: accessToken,
		client:      &http.Client{},
	}
}

type geniusSearchResponse struct {
	Response struct {
		Hits []struct {
			Result struct {
				Title         string `json:"title"`
				PrimaryArtist struct {
					Name string `json:"name"`
				} `json:"primary_artist"`
				URL string `json:"url"`
			} `json:"result"`
		} `json:"hits"`
	} `json:"response"`
}

// FetchLyrics retrieves plain text lyrics from Genius by searching and scraping.
func (p *GeniusProvider) FetchLyrics(artist, title string) (string, error) {
	query := fmt.Sprintf("%s %s", artist, title)
	searchURL := fmt.Sprintf("https://api.genius.com/search?q=%s", url.QueryEscape(query))

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create genius request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.accessToken)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("genius request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read genius response: %w", err)
	}

	var searchResp geniusSearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return "", fmt.Errorf("failed to parse genius response: %w", err)
	}

	if len(searchResp.Response.Hits) == 0 {
		return "", fmt.Errorf("song not found on genius")
	}

	songURL := searchResp.Response.Hits[0].Result.URL

	lyrics, err := p.scrapeLyrics(songURL)
	if err != nil {
		return "", fmt.Errorf("failed to scrape lyrics: %w", err)
	}

	return lyrics, nil
}

// FetchSynced is not supported by Genius (only plain text).
func (p *GeniusProvider) FetchSynced(artist, title string) ([]Line, error) {
	return nil, fmt.Errorf("genius does not provide synced lyrics")
}

func (p *GeniusProvider) scrapeLyrics(songURL string) (string, error) {
	resp, err := p.client.Get(songURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	html := string(body)

	re := regexp.MustCompile(`<div[^>]*data-lyrics-container="true"[^>]*>(.*?)</div>`)
	matches := re.FindAllStringSubmatch(html, -1)

	if len(matches) == 0 {
		return "", fmt.Errorf("lyrics not found in page")
	}

	var lyrics []string
	for _, match := range matches {
		if len(match) > 1 {
			lyricHTML := match[1]
			lyricHTML = regexp.MustCompile(`<br[^>]*>`).ReplaceAllString(lyricHTML, "\n")
			lyricHTML = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(lyricHTML, "")
			lyricHTML = strings.ReplaceAll(lyricHTML, "&amp;", "&")
			lyricHTML = strings.ReplaceAll(lyricHTML, "&quot;", "\"")
			lyricHTML = strings.ReplaceAll(lyricHTML, "&#x27;", "'")
			lyricHTML = strings.ReplaceAll(lyricHTML, "&lt;", "<")
			lyricHTML = strings.ReplaceAll(lyricHTML, "&gt;", ">")
			lyrics = append(lyrics, lyricHTML)
		}
	}

	if len(lyrics) == 0 {
		return "", fmt.Errorf("no lyrics extracted")
	}

	return strings.TrimSpace(strings.Join(lyrics, "\n\n")), nil
}
