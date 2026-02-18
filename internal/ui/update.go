package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"lyrics-tui/internal/lyrics"
	"lyrics-tui/internal/parse"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		return m.handleWindowSizeMsg(msg)

	case tickMsg:
		return m, tea.Batch(m.detectCurrentSong(), tickEverySecond())

	case positionTickMsg:
		if m.hasSyncedLyrics {
			return m, tea.Batch(m.getPlaybackPosition(), tickPosition())
		}
		return m, tickPosition()

	case playbackPosition:
		return m.handlePlaybackPosition(msg)

	case mprisData:
		return m.handleMPRISData(msg)

	case parsedResult:
		return m.handleParsedResult(msg)

	case searchResult:
		return m.handleSearchResult(msg)
	}

	if m.settingsOpen {
		var smCmd, saCmd tea.Cmd
		m.settingsModel, smCmd = m.settingsModel.Update(msg)
		m.settingsAPIKey, saCmd = m.settingsAPIKey.Update(msg)
		return m, tea.Batch(smCmd, saCmd)
	}

	if m.searchModalOpen {
		var tiCmd tea.Cmd
		m.input, tiCmd = m.input.Update(msg)
		return m, tiCmd
	}

	if m.cachedSongsModalOpen {
		var cfCmd tea.Cmd
		m.cachedSongsFilter, cfCmd = m.cachedSongsFilter.Update(msg)
		return m, cfCmd
	}

	m.viewport, _ = m.viewport.Update(msg)
	return m, nil
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.settingsOpen {
		return m.handleSettingsKeyMsg(msg)
	}
	if m.searchModalOpen {
		return m.handleSearchModalKeyMsg(msg)
	}
	if m.cachedSongsModalOpen {
		return m.handleCachedSongsKeyMsg(msg)
	}

	switch msg.String() {
	case "ctrl+c", "esc":
		return m, tea.Quit

	case "ctrl+o":
		return m.openSettings()

	case "/":
		m.searchModalOpen = true
		m.input.SetValue("")
		m.input.Focus()
		return m, nil

	case "ctrl+_":
		m.cachedSongs = m.lyricsService.ListAllCached()
		m.cachedSongsFiltered = m.cachedSongs
		m.cachedSongsCursor = 0
		m.cachedSongsFilter.SetValue("")
		m.cachedSongsFilter.Focus()
		m.cachedSongsModalOpen = true
		return m, nil

	case "tab":
		m.autoDetectMode = !m.autoDetectMode
		if m.autoDetectMode {
			return m, tea.Batch(m.detectCurrentSong(), tickEverySecond())
		}
		return m, nil

	case "f":
		m.followMode = !m.followMode
		return m, nil

	case "+", "=":
		if m.hasSyncedLyrics {
			m.offset += 0.1
			m.viewport.SetContent(m.renderSyncedLyrics())
			if m.mprisArtist != "" && m.mprisTitle != "" {
				go m.saveOffsetToCache()
			}
		}

	case "-", "_":
		if m.hasSyncedLyrics {
			m.offset -= 0.1
			m.viewport.SetContent(m.renderSyncedLyrics())
			if m.mprisArtist != "" && m.mprisTitle != "" {
				go m.saveOffsetToCache()
			}
		}

	case "up", "k":
		m.viewport, _ = m.viewport.Update(msg)
	case "down", "j":
		m.viewport, _ = m.viewport.Update(msg)
	}

	return m, nil
}

// --- search modal ---

func (m Model) handleSearchModalKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.searchModalOpen = false
		m.input.Blur()
		return m, nil
	case "enter":
		query := m.input.Value()
		if query == "" {
			return m, nil
		}
		m.searchModalOpen = false
		m.input.Blur()
		m.searching = true
		m.viewport.SetContent("Searching...")
		return m, m.searchLyrics(query)
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// --- cached songs modal ---

func (m Model) handleCachedSongsKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.cachedSongsModalOpen = false
		m.cachedSongsFilter.Blur()
		return m, nil
	case "up":
		if m.cachedSongsCursor > 0 {
			m.cachedSongsCursor--
		}
		return m, nil
	case "down":
		if m.cachedSongsCursor < len(m.cachedSongsFiltered)-1 {
			m.cachedSongsCursor++
		}
		return m, nil
	case "enter":
		if len(m.cachedSongsFiltered) == 0 {
			return m, nil
		}
		entry := m.cachedSongsFiltered[m.cachedSongsCursor]
		cached, err := m.lyricsService.LoadFromCache(entry.Artist, entry.Title)
		if err != nil {
			m.cachedSongsModalOpen = false
			m.cachedSongsFilter.Blur()
			m.viewport.SetContent(fmt.Sprintf("Error loading from cache: %s", err))
			return m, nil
		}

		m.artist = cached.Artist
		m.title = cached.Title
		m.lyrics = cached.Lyrics
		m.syncedLyrics = cached.SyncedLyrics
		m.hasSyncedLyrics = cached.HasSyncedLyrics
		m.offset = cached.Offset
		m.parsedArtist = cached.Artist
		m.parsedTitle = cached.Title
		m.playbackPosition = 0
		m.searching = false

		if m.hasSyncedLyrics {
			m.viewport.SetContent(m.renderSyncedLyrics())
		} else {
			m.viewport.SetContent(cached.Lyrics)
		}

		m.cachedSongsModalOpen = false
		m.cachedSongsFilter.Blur()
		return m, nil
	}

	var cmd tea.Cmd
	m.cachedSongsFilter, cmd = m.cachedSongsFilter.Update(msg)
	m.cachedSongsFiltered = m.filterCachedSongs()
	m.cachedSongsCursor = 0
	return m, cmd
}

func (m Model) filterCachedSongs() []lyrics.CachedSongEntry {
	query := strings.ToLower(m.cachedSongsFilter.Value())
	if query == "" {
		return m.cachedSongs
	}
	var filtered []lyrics.CachedSongEntry
	for _, entry := range m.cachedSongs {
		haystack := strings.ToLower(entry.Artist + " " + entry.Title)
		if strings.Contains(haystack, query) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// --- settings modal ---

func (m Model) openSettings() (tea.Model, tea.Cmd) {
	m.settingsOpen = true
	m.settingsCursor = 0

	for i, id := range parse.AllProviders {
		if id == parse.ProviderID(m.config.Provider) {
			m.settingsProviderIdx = i
			break
		}
	}

	m.settingsModel.SetValue(m.config.Model)
	m.settingsAPIKey.SetValue(m.config.APIKey)
	m.settingsModel.Blur()
	m.settingsAPIKey.Blur()

	return m, nil
}

func (m Model) settingsMaxField() int {
	providerID := parse.AllProviders[m.settingsProviderIdx]
	if providerID != parse.ProviderOllama {
		return 3
	}
	return 2
}

func (m Model) settingsClearCacheField() int {
	return m.settingsMaxField()
}

func (m Model) handleSettingsKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxField := m.settingsMaxField()

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "esc", "ctrl+o":
		m.settingsOpen = false
		return m, nil

	case "enter":
		if m.settingsCursor == m.settingsClearCacheField() {
			m.lyricsService.ClearCache()
			return m, nil
		}

		m.config.Provider = string(parse.AllProviders[m.settingsProviderIdx])
		m.config.Model = m.settingsModel.Value()
		m.config.APIKey = m.settingsAPIKey.Value()
		m.config.Save()

		newParser, err := parse.NewProviderFromConfig(m.config)
		if err == nil {
			m.parser = newParser
		}

		m.settingsOpen = false
		return m, nil

	case "tab", "down":
		m.settingsCursor++
		if m.settingsCursor > maxField {
			m.settingsCursor = 0
		}
		m = m.focusSettingsField()
		return m, nil

	case "shift+tab", "up":
		m.settingsCursor--
		if m.settingsCursor < 0 {
			m.settingsCursor = maxField
		}
		m = m.focusSettingsField()
		return m, nil

	case "left":
		if m.settingsCursor == 0 {
			m.settingsProviderIdx--
			if m.settingsProviderIdx < 0 {
				m.settingsProviderIdx = len(parse.AllProviders) - 1
			}
			newID := parse.AllProviders[m.settingsProviderIdx]
			m.settingsModel.SetValue(parse.DefaultModelForProvider(newID))
			m.settingsAPIKey.SetValue("")
			newMax := m.settingsMaxField()
			if m.settingsCursor > newMax {
				m.settingsCursor = newMax
			}
			return m, nil
		}

	case "right":
		if m.settingsCursor == 0 {
			m.settingsProviderIdx++
			if m.settingsProviderIdx >= len(parse.AllProviders) {
				m.settingsProviderIdx = 0
			}
			newID := parse.AllProviders[m.settingsProviderIdx]
			m.settingsModel.SetValue(parse.DefaultModelForProvider(newID))
			m.settingsAPIKey.SetValue("")
			newMax := m.settingsMaxField()
			if m.settingsCursor > newMax {
				m.settingsCursor = newMax
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	switch m.settingsCursor {
	case 1:
		m.settingsModel, cmd = m.settingsModel.Update(msg)
	case 2:
		m.settingsAPIKey, cmd = m.settingsAPIKey.Update(msg)
	}
	return m, cmd
}

func (m Model) focusSettingsField() Model {
	m.settingsModel.Blur()
	m.settingsAPIKey.Blur()
	switch m.settingsCursor {
	case 1:
		m.settingsModel.Focus()
	case 2:
		m.settingsAPIKey.Focus()
	}
	return m
}

// --- window / playback / mpris handlers ---

func (m Model) handleWindowSizeMsg(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height

	leftWidth := m.width / 4
	rightWidth := m.width - leftWidth - 6

	if !m.ready {
		m.viewport.Width = rightWidth
		m.viewport.Height = msg.Height - 4
		m.ready = true
	} else {
		m.viewport.Width = rightWidth
		m.viewport.Height = msg.Height - 4
	}

	return m, nil
}

func (m Model) handlePlaybackPosition(msg playbackPosition) (tea.Model, tea.Cmd) {
	if msg.err != nil || m.searching {
		return m, nil
	}

	if time.Now().Before(m.ignorePositionUntil) {
		return m, nil
	}

	if msg.position >= msg.duration && msg.duration > 0 && m.artist != "" {
		return m, tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
			return m.getPlaybackPosition()()
		})
	}

	m.playbackPosition = msg.position
	m.duration = msg.duration

	if m.hasSyncedLyrics {
		oldYOffset := m.viewport.YOffset
		m.viewport.SetContent(m.renderSyncedLyrics())

		if m.followMode {
			currentIdx := m.getCurrentLineIndex()
			if currentIdx >= 0 {
				centerOffset := currentIdx - (m.viewport.Height / 2)
				if centerOffset < 0 {
					centerOffset = 0
				}
				m.viewport.SetYOffset(centerOffset)
			}
		} else {
			m.viewport.SetYOffset(oldYOffset)
		}
	}

	return m, nil
}

func (m Model) handleMPRISData(msg mprisData) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.debugInfo = fmt.Sprintf("Error: %v", msg.err)
		return m, nil
	}

	if msg.artist == "" || msg.title == "" {
		m.debugInfo = ""
		m.mprisArtist = ""
		m.mprisTitle = ""
		return m, nil
	}

	m.debugInfo = fmt.Sprintf("%s\n%s", msg.artist, msg.title)
	m.mprisArtist = msg.artist
	m.mprisTitle = msg.title

	if !m.autoDetectMode {
		return m, nil
	}

	songKey := msg.artist + " - " + msg.title
	if songKey == m.lastDetectedSong {
		return m, nil
	}

	m.lastDetectedSong = songKey
	m.searching = true

	m.artist = ""
	m.title = ""
	m.lyrics = ""
	m.syncedLyrics = nil
	m.hasSyncedLyrics = false
	m.playbackPosition = 0
	m.duration = 0
	m.parsedArtist = ""
	m.parsedTitle = ""
	m.ignorePositionUntil = time.Now().Add(2 * time.Second)

	cached, err := m.lyricsService.LoadFromCache(msg.artist, msg.title)
	if err == nil {
		m.searching = false
		m.artist = cached.Artist
		m.title = cached.Title
		m.lyrics = cached.Lyrics
		m.syncedLyrics = cached.SyncedLyrics
		m.hasSyncedLyrics = cached.HasSyncedLyrics
		m.parsedArtist = cached.Artist
		m.parsedTitle = cached.Title
		m.playbackPosition = 0
		m.offset = cached.Offset
		m.ignorePositionUntil = time.Now().Add(1 * time.Second)

		if m.hasSyncedLyrics {
			m.viewport.SetContent(m.renderSyncedLyrics())
		} else {
			m.viewport.SetContent(cached.Lyrics)
		}
		return m, tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
			return m.getPlaybackPosition()()
		})
	}

	query := msg.artist + " " + msg.title
	m.viewport.SetContent(fmt.Sprintf("New song detected!\n\n%s\n\nFetching lyrics...", query))
	return m, m.searchLyricsWithMpris(query, msg.artist, msg.title)
}

func (m Model) handleParsedResult(msg parsedResult) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.searching = false
		m.err = msg.err
		m.viewport.SetContent(fmt.Sprintf("Parse error: %s", msg.err))
		return m, nil
	}

	m.parsedArtist = msg.artist
	m.parsedTitle = msg.title
	m.viewport.SetContent(fmt.Sprintf("Parsed!\nFetching lyrics for:\n%s - %s", msg.artist, msg.title))
	return m, m.fetchLyrics(msg.artist, msg.title, msg.mprisArtist, msg.mprisTitle)
}

func (m Model) handleSearchResult(msg searchResult) (tea.Model, tea.Cmd) {
	m.searching = false

	if msg.err != nil {
		m.err = msg.err
		m.viewport.SetContent(fmt.Sprintf("Error: %s", msg.err))
		return m, nil
	}

	m.artist = msg.song.Artist
	m.title = msg.song.Title
	m.lyrics = msg.song.Lyrics
	m.syncedLyrics = msg.song.SyncedLyrics
	m.hasSyncedLyrics = msg.song.HasSyncedLyrics
	m.playbackPosition = 0
	m.offset = 0
	m.ignorePositionUntil = time.Now().Add(1 * time.Second)

	if msg.mprisArtist != "" && msg.mprisTitle != "" {
		m.lyricsService.SaveToCache(msg.mprisArtist, msg.mprisTitle, msg.song, 0)
	}

	if m.hasSyncedLyrics {
		m.viewport.SetContent(m.renderSyncedLyrics())
	} else {
		m.viewport.SetContent(msg.song.Lyrics)
	}

	return m, tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return m.getPlaybackPosition()()
	})
}

func (m Model) saveOffsetToCache() {
	if m.mprisArtist == "" || m.mprisTitle == "" {
		return
	}
	m.lyricsService.UpdateOffset(m.mprisArtist, m.mprisTitle, m.offset)
}
