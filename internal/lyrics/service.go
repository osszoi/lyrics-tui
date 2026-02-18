package lyrics

import "fmt"

// Service coordinates lyrics fetching from multiple providers with caching.
type Service struct {
	syncedProvider Provider
	lyricsProvider Provider
	cache          *Cache
}

// NewService creates a new lyrics service.
// syncedProvider is tried first for time-synced lyrics.
// lyricsProvider is used as fallback for plain lyrics.
func NewService(syncedProvider, lyricsProvider Provider, cache *Cache) *Service {
	return &Service{
		syncedProvider: syncedProvider,
		lyricsProvider: lyricsProvider,
		cache:          cache,
	}
}

// Fetch retrieves lyrics, trying cache first, then providers.
// Priority: cache -> synced lyrics -> plain lyrics.
func (s *Service) Fetch(artist, title string) (*Song, error) {
	cached, err := s.cache.Load(artist, title)
	if err == nil {
		return &Song{
			Artist:          cached.Artist,
			Title:           cached.Title,
			Lyrics:          cached.Lyrics,
			SyncedLyrics:    cached.SyncedLyrics,
			HasSyncedLyrics: cached.HasSyncedLyrics,
		}, nil
	}

	syncedLyrics, err := s.syncedProvider.FetchSynced(artist, title)
	if err == nil && len(syncedLyrics) > 0 {
		song := &Song{
			Artist:          artist,
			Title:           title,
			SyncedLyrics:    syncedLyrics,
			HasSyncedLyrics: true,
		}
		s.saveToCache(artist, title, song, 0)
		return song, nil
	}

	plainLyrics, err := s.lyricsProvider.FetchLyrics(artist, title)
	if err != nil {
		return nil, fmt.Errorf("all providers failed: %w", err)
	}

	song := &Song{
		Artist:          artist,
		Title:           title,
		Lyrics:          plainLyrics,
		HasSyncedLyrics: false,
	}
	s.saveToCache(artist, title, song, 0)
	return song, nil
}

// LoadFromCache retrieves a song from cache, including offset.
func (s *Service) LoadFromCache(artist, title string) (*CachedSong, error) {
	return s.cache.Load(artist, title)
}

// SaveToCache stores a song in cache.
func (s *Service) SaveToCache(artist, title string, song *Song, offset float64) error {
	return s.saveToCache(artist, title, song, offset)
}

// UpdateOffset updates the timing offset for cached lyrics.
func (s *Service) UpdateOffset(artist, title string, offset float64) error {
	return s.cache.UpdateOffset(artist, title, offset)
}

func (s *Service) ListAllCached() []CachedSongEntry {
	return s.cache.ListAll()
}

func (s *Service) ClearCache() error {
	return s.cache.ClearAll()
}

func (s *Service) CachedSongCount() int {
	return s.cache.Count()
}

func (s *Service) saveToCache(artist, title string, song *Song, offset float64) error {
	cached := &CachedSong{
		Artist:          artist,
		Title:           title,
		Lyrics:          song.Lyrics,
		SyncedLyrics:    song.SyncedLyrics,
		HasSyncedLyrics: song.HasSyncedLyrics,
		Offset:          offset,
	}
	return s.cache.Save(cached)
}
