package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles all messages and updates the model state.
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

	if !m.autoDetectMode {
		m.input, tiCmd = m.input.Update(msg)
	}
	m.viewport, vpCmd = m.viewport.Update(msg)

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc":
		return m, tea.Quit

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
		m.debugInfo = m.debugInfo + "\n\n✓ Loaded from cache (skipped LLM)!"
		return m, tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
			return m.getPlaybackPosition()()
		})
	}

	query := msg.artist + " " + msg.title
	m.viewport.SetContent(fmt.Sprintf("🔄 New song detected!\n\n%s\n\nFetching lyrics...", query))
	m.debugInfo = m.debugInfo + fmt.Sprintf("\n\nSearching: %s", query)
	return m, m.searchLyricsWithMpris(query, msg.artist, msg.title)
}

func (m Model) handleParsedResult(msg parsedResult) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.searching = false
		m.err = msg.err
		m.viewport.SetContent(fmt.Sprintf("❌ Parse error: %s", msg.err))
		m.debugInfo = m.debugInfo + "\n\n❌ Parse error"
		return m, nil
	}

	m.parsedArtist = msg.artist
	m.parsedTitle = msg.title
	m.viewport.SetContent(fmt.Sprintf("✓ Parsed!\nFetching lyrics for:\n%s - %s", msg.artist, msg.title))
	return m, m.fetchLyrics(msg.artist, msg.title, msg.mprisArtist, msg.mprisTitle)
}

func (m Model) handleSearchResult(msg searchResult) (tea.Model, tea.Cmd) {
	m.searching = false

	if msg.err != nil {
		m.err = msg.err
		m.viewport.SetContent(fmt.Sprintf("❌ Error: %s", msg.err))
		m.debugInfo = m.debugInfo + "\n\n❌ Error: " + msg.err.Error()
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
	m.debugInfo = m.debugInfo + "\n\n✓ Lyrics loaded!"

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
