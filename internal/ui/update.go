package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"lyrics-tui/internal/parse"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

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

	if !m.autoDetectMode {
		m.input, tiCmd = m.input.Update(msg)
	}
	m.viewport, vpCmd = m.viewport.Update(msg)

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.settingsOpen {
		return m.handleSettingsKeyMsg(msg)
	}

	switch msg.String() {
	case "ctrl+c", "esc":
		return m, tea.Quit

	case "ctrl+o":
		return m.openSettings()

	case "tab":
		m.autoDetectMode = !m.autoDetectMode
		if m.autoDetectMode {
			m.input.Blur()
			return m, tea.Batch(m.detectCurrentSong(), tickEverySecond())
		}
		m.input.Focus()
		return m, nil

	case "enter":
		if !m.autoDetectMode && !m.searching && m.input.Value() != "" {
			query := m.input.Value()
			m.input.SetValue("")
			m.searching = true
			m.viewport.SetContent("Searching...")
			return m, m.searchLyrics(query)
		}

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

	case "/", "?":
		m.followMode = !m.followMode
	}

	var tiCmd tea.Cmd
	var vpCmd tea.Cmd
	if !m.autoDetectMode {
		m.input, tiCmd = m.input.Update(msg)
	}
	m.viewport, vpCmd = m.viewport.Update(msg)

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m Model) openSettings() (tea.Model, tea.Cmd) {
	m.settingsOpen = true
	m.settingsCursor = 0
	m.input.Blur()

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

func (m Model) handleSettingsKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	providerID := parse.AllProviders[m.settingsProviderIdx]
	maxField := 1
	if providerID != parse.ProviderOllama {
		maxField = 2
	}

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "esc", "ctrl+o":
		m.settingsOpen = false
		if !m.autoDetectMode {
			m.input.Focus()
		}
		return m, nil

	case "enter":
		m.config.Provider = string(parse.AllProviders[m.settingsProviderIdx])
		m.config.Model = m.settingsModel.Value()
		m.config.APIKey = m.settingsAPIKey.Value()
		m.config.Save()

		newParser, err := parse.NewProviderFromConfig(m.config)
		if err == nil {
			m.parser = newParser
		}

		m.settingsOpen = false
		if !m.autoDetectMode {
			m.input.Focus()
		}
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
			newMax := 1
			if newID != parse.ProviderOllama {
				newMax = 2
			}
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
			newMax := 1
			if newID != parse.ProviderOllama {
				newMax = 2
			}
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

func (m Model) handleWindowSizeMsg(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height

	leftWidth := m.width / 4
	rightWidth := m.width - leftWidth - 6

	m.input.Width = leftWidth - 6

	if !m.ready {
		m.viewport = m.viewport
		m.viewport.Width = rightWidth
		m.viewport.Height = msg.Height - 6
		m.ready = true
	} else {
		m.viewport.Width = rightWidth
		m.viewport.Height = msg.Height - 6
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
				targetLine := currentIdx
				viewportHeight := m.viewport.Height
				centerOffset := targetLine - (viewportHeight / 2)
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
		m.debugInfo = fmt.Sprintf("No data\nArtist: '%s'\nTitle: '%s'", msg.artist, msg.title)
		return m, nil
	}

	m.debugInfo = fmt.Sprintf("%s\n%s", msg.artist, msg.title)

	if !m.autoDetectMode {
		return m, nil
	}

	songKey := msg.artist + " - " + msg.title
	if songKey == m.lastDetectedSong {
		return m, nil
	}

	m.lastDetectedSong = songKey
	m.mprisArtist = msg.artist
	m.mprisTitle = msg.title
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
		m.debugInfo = m.debugInfo + "\n\nLoaded from cache (skipped LLM)!"
		return m, tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
			return m.getPlaybackPosition()()
		})
	}

	query := msg.artist + " " + msg.title
	m.viewport.SetContent(fmt.Sprintf("New song detected!\n\n%s\n\nFetching lyrics...", query))
	m.debugInfo = m.debugInfo + fmt.Sprintf("\n\nSearching: %s", query)
	return m, m.searchLyricsWithMpris(query, msg.artist, msg.title)
}

func (m Model) handleParsedResult(msg parsedResult) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.searching = false
		m.err = msg.err
		m.viewport.SetContent(fmt.Sprintf("Parse error: %s", msg.err))
		m.debugInfo = m.debugInfo + "\n\nParse error"
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
		m.debugInfo = m.debugInfo + "\n\nError: " + msg.err.Error()
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
	m.debugInfo = m.debugInfo + "\n\nLyrics loaded!"

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
