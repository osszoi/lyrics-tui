package lyrics

// Line represents a single line of synced lyrics with its timestamp.
type Line struct {
	Timestamp float64 `json:"timestamp"`
	Text      string  `json:"text"`
}

// Song contains lyrics information for a song.
type Song struct {
	Artist          string
	Title           string
	Lyrics          string
	SyncedLyrics    []Line
	HasSyncedLyrics bool
}

// Provider defines the interface for lyrics sources.
type Provider interface {
	// FetchLyrics retrieves plain text lyrics for a song.
	FetchLyrics(artist, title string) (string, error)

	// FetchSynced retrieves time-synced lyrics for a song.
	// Returns nil slice if synced lyrics are not available.
	FetchSynced(artist, title string) ([]Line, error)
}

type CachedSongEntry struct {
	Artist string
	Title  string
}

// CachedSong represents a song stored in cache with offset information.
type CachedSong struct {
	Artist          string  `json:"artist"`
	Title           string  `json:"title"`
	Lyrics          string  `json:"lyrics"`
	SyncedLyrics    []Line  `json:"syncedLyrics"`
	HasSyncedLyrics bool    `json:"hasSyncedLyrics"`
	Offset          float64 `json:"offset"`
}
