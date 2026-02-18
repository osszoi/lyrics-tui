package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"lyrics-tui/internal/parse"
)

func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	if m.settingsOpen {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.renderSettingsModal())
	}
	if m.searchModalOpen {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.renderSearchModal())
	}
	if m.cachedSongsModalOpen {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.renderCachedSongsModal())
	}

	leftWidth := m.width / 4
	rightWidth := m.width - leftWidth - 6

	leftColumn := m.renderLeftColumn(leftWidth)
	lyricsBox := m.renderLyricsBox(rightWidth)

	content := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, lyricsBox)

	help := helpStyle.Render("\n/: search • Ctrl+/: cached • Tab: auto-detect • f: follow • +/-: timing • Ctrl+O: settings • Esc: quit")

	return lipgloss.JoinVertical(lipgloss.Left, content, help)
}

// --- left panel: 3 bordered boxes ---

func (m Model) renderLeftColumn(width int) string {
	totalHeight := m.height - 2
	innerWidth := width - 2

	box1Height := 2
	remaining := totalHeight - (box1Height + 2)
	box2Height := remaining / 2
	box3Height := remaining - box2Height

	box1 := m.renderHeaderBox(innerWidth, box1Height)
	box2 := m.renderLoadedSongBox(innerWidth, box2Height-2)
	box3 := m.renderNowPlayingBox(innerWidth, box3Height-2)

	return lipgloss.JoinVertical(lipgloss.Left, box1, box2, box3)
}

func (m Model) renderHeaderBox(width, height int) string {
	versionStr := m.version
	if versionStr == "dev" {
		versionStr = "(dev)"
	}

	line1 := titleStyle.Render(fmt.Sprintf("♪ Lyrics TUI %s", versionStr))
	line2 := helpStyle.Render(fmt.Sprintf("  %s (%s)", m.parser.Name(), m.config.Model))

	content := lipgloss.JoinVertical(lipgloss.Left, line1, line2)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lavender).
		Width(width).
		Height(height).
		Render(content)
}

func (m Model) renderLoadedSongBox(width, height int) string {
	var parts []string

	parts = append(parts, helpStyle.Render("Loaded Song"))
	parts = append(parts, "")

	if m.artist != "" && m.title != "" {
		parts = append(parts, infoStyle.Render("♪ "+m.artist))
		parts = append(parts, infoStyle.Render("  "+m.title))

		if m.searching {
			parts = append(parts, "")
			parts = append(parts, activeStyle.Render("Searching..."))
		}

		if !m.hasSyncedLyrics && m.lyrics != "" {
			parts = append(parts, "")
			parts = append(parts, warningStyle.Render("No synced lyrics"))
		}

		if m.hasSyncedLyrics && m.offset != 0 {
			parts = append(parts, "")
			parts = append(parts, helpStyle.Render(fmt.Sprintf("Offset: %+.1fs", m.offset)))
		}

		if m.hasSyncedLyrics {
			followStr := "ON"
			followColor := activeStyle
			if !m.followMode {
				followStr = "OFF"
				followColor = warningStyle
			}
			parts = append(parts, followColor.Render("Follow: "+followStr))
		}
	} else {
		parts = append(parts, helpStyle.Render("No song loaded"))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lavender).
		Width(width).
		Height(height).
		Render(content)
}

func (m Model) renderNowPlayingBox(width, height int) string {
	var parts []string

	parts = append(parts, helpStyle.Render("Now Playing"))

	if m.autoDetectMode {
		parts = append(parts, activeStyle.Render("Auto-detect: ON"))
	} else {
		parts = append(parts, helpStyle.Render("Auto-detect: OFF"))
	}

	parts = append(parts, "")

	if m.mprisArtist != "" && m.mprisTitle != "" {
		parts = append(parts, infoStyle.Render("♪ "+m.mprisArtist))
		parts = append(parts, infoStyle.Render("  "+m.mprisTitle))
		parts = append(parts, "")
		parts = append(parts, m.renderProgressBar(width-2))
	} else {
		parts = append(parts, helpStyle.Render("Nothing detected"))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lavender).
		Width(width).
		Height(height).
		Render(content)
}

// --- modals ---

func (m Model) renderSearchModal() string {
	var parts []string

	parts = append(parts, titleStyle.Render("Search"))
	parts = append(parts, "")
	parts = append(parts, m.input.View())
	parts = append(parts, "")
	parts = append(parts, helpStyle.Render("Enter: search · Esc: cancel"))

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(mauve).
		Padding(1, 2).
		Width(50).
		Render(content)
}

func (m Model) renderCachedSongsModal() string {
	var parts []string

	parts = append(parts, titleStyle.Render("Cached Songs"))
	parts = append(parts, "")
	parts = append(parts, m.cachedSongsFilter.View())
	parts = append(parts, "")

	filtered := m.cachedSongsFiltered
	if len(filtered) == 0 {
		if m.cachedSongsFilter.Value() != "" {
			parts = append(parts, helpStyle.Render("No matches"))
		} else {
			parts = append(parts, helpStyle.Render("No cached songs found"))
		}
	} else {
		maxVisible := 15
		start := 0
		if m.cachedSongsCursor >= maxVisible {
			start = m.cachedSongsCursor - maxVisible + 1
		}
		end := start + maxVisible
		if end > len(filtered) {
			end = len(filtered)
		}

		for i := start; i < end; i++ {
			entry := filtered[i]
			line := fmt.Sprintf("%s - %s", entry.Title, entry.Artist)
			if i == m.cachedSongsCursor {
				parts = append(parts, activeStyle.Render("> "+line))
			} else {
				parts = append(parts, helpStyle.Render("  "+line))
			}
		}

		if len(filtered) > maxVisible {
			parts = append(parts, "")
			parts = append(parts, helpStyle.Render(fmt.Sprintf("  %d/%d", m.cachedSongsCursor+1, len(filtered))))
		}
	}

	parts = append(parts, "")
	parts = append(parts, helpStyle.Render("Enter: load · Esc: cancel · ↑/↓: navigate"))

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(mauve).
		Padding(1, 2).
		Width(60).
		Render(content)
}

func (m Model) renderSettingsModal() string {
	width := 50
	var parts []string

	parts = append(parts, titleStyle.Render("Settings"))
	parts = append(parts, "")

	providerID := parse.AllProviders[m.settingsProviderIdx]
	providerName := parse.ProviderName(providerID)

	if m.settingsCursor == 0 {
		parts = append(parts, activeStyle.Render("> ")+"Provider    "+infoStyle.Render("◂ "+providerName+" ▸"))
	} else {
		parts = append(parts, "  Provider    "+infoStyle.Render("  "+providerName))
	}
	parts = append(parts, "")

	if m.settingsCursor == 1 {
		parts = append(parts, activeStyle.Render("> ")+"Model       "+m.settingsModel.View())
	} else {
		parts = append(parts, "  Model       "+m.settingsModel.View())
	}
	parts = append(parts, "")

	if providerID != parse.ProviderOllama {
		if m.settingsCursor == 2 {
			parts = append(parts, activeStyle.Render("> ")+"API Key     "+m.settingsAPIKey.View())
		} else {
			parts = append(parts, "  API Key     "+m.settingsAPIKey.View())
		}

		envVar := ""
		switch providerID {
		case parse.ProviderOpenAI:
			envVar = "OPENAI_API_KEY"
		case parse.ProviderGemini:
			envVar = "GEMINI_API_KEY"
		}
		if envVar != "" {
			parts = append(parts, helpStyle.Render("                default: "+envVar))
		}
		parts = append(parts, "")
	}

	aiLyricsField := m.settingsAILyricsField()
	aiLyricsVal := "OFF"
	if m.settingsAILyrics {
		aiLyricsVal = "ON"
	}
	if m.settingsCursor == aiLyricsField {
		parts = append(parts, activeStyle.Render("> ")+"AI Lyrics   "+infoStyle.Render("◂ "+aiLyricsVal+" ▸"))
	} else {
		parts = append(parts, "  AI Lyrics   "+infoStyle.Render("  "+aiLyricsVal))
	}
	parts = append(parts, "")

	clearCacheField := m.settingsClearCacheField()
	count := m.lyricsService.CachedSongCount()
	clearLabel := fmt.Sprintf("Clear Cache (%d songs)", count)
	if m.settingsCursor == clearCacheField {
		parts = append(parts, errorStyle.Render("> "+clearLabel))
	} else {
		parts = append(parts, helpStyle.Render("  "+clearLabel))
	}
	parts = append(parts, "")

	parts = append(parts, helpStyle.Render("  Enter: save · Esc: cancel"))
	parts = append(parts, helpStyle.Render("  ◂/▸: change provider · Tab: next field"))

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(mauve).
		Padding(1, 2).
		Width(width).
		Render(content)
}

// --- lyrics rendering ---

func (m Model) renderLyricsBox(width int) string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lavender).
		Width(width).
		Height(m.height - 4).
		Render(m.viewport.View())
}

func (m Model) renderProgressBar(width int) string {
	if m.duration == 0 {
		barWidth := width - 14
		if barWidth < 5 {
			barWidth = 5
		}
		bar := strings.Repeat("░", barWidth)
		return helpStyle.Render(fmt.Sprintf("%s 0:00 / 0:00", bar))
	}

	barWidth := width - 14
	if barWidth < 5 {
		barWidth = 5
	}
	progress := m.playbackPosition / m.duration
	if progress > 1 {
		progress = 1
	}

	filled := int(progress * float64(barWidth))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	currentTime := formatTime(m.playbackPosition)
	totalTime := formatTime(m.duration)

	return infoStyle.Render(fmt.Sprintf("%s %s / %s", bar, currentTime, totalTime))
}

func (m Model) renderSyncedLyrics() string {
	if len(m.syncedLyrics) == 0 {
		return "No lyrics available"
	}

	var rendered []string

	if !m.followMode {
		grayStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#9399b2"))
		for _, line := range m.syncedLyrics {
			rendered = append(rendered, grayStyle.Render("  "+line.Text))
		}
		return strings.Join(rendered, "\n")
	}

	currentIdx := m.getCurrentLineIndex()

	dimmedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#45475a"))
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#cdd6f4"))

	for i, line := range m.syncedLyrics {
		if i == currentIdx {
			rendered = append(rendered, normalStyle.Render("► "+line.Text))
		} else {
			rendered = append(rendered, dimmedStyle.Render("  "+line.Text))
		}
	}

	return strings.Join(rendered, "\n")
}

func (m Model) getCurrentLineIndex() int {
	if len(m.syncedLyrics) == 0 {
		return -1
	}

	adjustedPosition := m.playbackPosition - m.offset
	currentIdx := -1
	for i, line := range m.syncedLyrics {
		if line.Timestamp <= adjustedPosition {
			currentIdx = i
		}
	}
	return currentIdx
}

func formatTime(seconds float64) string {
	mins := int(seconds) / 60
	secs := int(seconds) % 60
	return fmt.Sprintf("%d:%02d", mins, secs)
}
