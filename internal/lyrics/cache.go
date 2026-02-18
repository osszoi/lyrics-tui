package lyrics

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Cache provides persistent storage for fetched lyrics.
type Cache struct {
	dir string
}

// NewCache creates a new lyrics cache in the specified directory.
func NewCache(dir string) *Cache {
	return &Cache{dir: dir}
}

// Load retrieves cached lyrics for a song.
func (c *Cache) Load(artist, title string) (*CachedSong, error) {
	path := c.cachePath(artist, title)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cache miss: %w", err)
	}

	var cached CachedSong
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, fmt.Errorf("failed to parse cache: %w", err)
	}

	return &cached, nil
}

// Save stores lyrics in the cache.
func (c *Cache) Save(song *CachedSong) error {
	if err := os.MkdirAll(c.dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache dir: %w", err)
	}

	data, err := json.MarshalIndent(song, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	path := c.cachePath(song.Artist, song.Title)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache: %w", err)
	}

	return nil
}

// UpdateOffset updates only the offset for a cached song.
func (c *Cache) UpdateOffset(artist, title string, offset float64) error {
	cached, err := c.Load(artist, title)
	if err != nil {
		return err
	}

	cached.Offset = offset
	return c.Save(cached)
}

func (c *Cache) cachePath(artist, title string) string {
	safeArtist := sanitizeFilename(artist)
	safeTitle := sanitizeFilename(title)
	return fmt.Sprintf("%s/%s_%s.json", c.dir, safeArtist, safeTitle)
}

func (c *Cache) ListAll() []CachedSongEntry {
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return nil
	}
	var songs []CachedSongEntry
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(c.dir, entry.Name()))
		if err != nil {
			continue
		}
		var cached CachedSong
		if err := json.Unmarshal(data, &cached); err != nil {
			continue
		}
		songs = append(songs, CachedSongEntry{Artist: cached.Artist, Title: cached.Title})
	}
	return songs
}

func sanitizeFilename(s string) string {
	s = strings.ToLower(s)
	s = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(s, "_")
	s = strings.Trim(s, "_")
	return s
}
