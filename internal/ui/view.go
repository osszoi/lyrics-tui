package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the TUI.
func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	leftWidth := m.width / 4
	rightWidth := m.width - leftWidth - 6

	leftColumn := m.renderLeftColumn(leftWidth)
	lyricsBox := m.renderLyricsBox(rightWidth)

	content := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, lyricsBox)

	help := helpStyle.Render("\nTab: auto-detect • Enter: search • +/-: adjust timing • /: follow mode • ↑/↓ or j/k: scroll • Esc: quit")

	return lipgloss.JoinVertical(lipgloss.Left, content, help)
}

func (m Model) renderLeftColumn(width int) string {
	var parts []string

	parts = append(parts, titleStyle.Render("♪ Lyrics TUI"))
	parts = append(parts, "")

	var inputContent string
	var inputBorderColor lipgloss.Color
	if m.autoDetectMode {
		inputContent = warningStyle.Render("🔍 Auto-detecting...")
		inputBorderColor = yellow
	} else {
		inputContent = m.input.View()
		inputBorderColor = lavender
	}

	inputBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(inputBorderColor).
		Width(width - 4).
		Padding(0, 1).
		Render(inputContent)
	parts = append(parts, inputBox)
	parts = append(parts, "")

	if m.artist != "" && m.title != "" {
		parts = append(parts, infoStyle.Render(fmt.Sprintf("🎵 %s", m.artist)))
		parts = append(parts, infoStyle.Render(fmt.Sprintf("   %s", m.title)))
		parts = append(parts, "")

		if m.duration > 0 {
			parts = append(parts, m.renderProgressBar())
			parts = append(parts, "")
		}

		if m.hasSyncedLyrics && m.offset != 0 {
			offsetStr := fmt.Sprintf("%+.1fs", m.offset)
			parts = append(parts, helpStyle.Render(fmt.Sprintf("Offset: %s", offsetStr)))
		}

		if m.hasSyncedLyrics && !m.followMode {
			parts = append(parts, warningStyle.Render("⏸ Follow: OFF"))
		}

		if m.hasSyncedLyrics && (m.offset != 0 || !m.followMode) {
			parts = append(parts, "")
		}

		if !m.hasSyncedLyrics && m.lyrics != "" {
			parts = append(parts, warningStyle.Render("⚠️ No synced lyrics"))
			parts = append(parts, helpStyle.Render("(using fallback)"))
		}
	} else {
		parts = append(parts, helpStyle.Render("No song loaded"))
	}

	if m.searching {
		parts = append(parts, "")
		parts = append(parts, activeStyle.Render("⏳ Searching..."))
	}

	leftContent := lipgloss.JoinVertical(lipgloss.Left, parts...)

	var footerParts []string
	footerParts = append(footerParts, "\n---")
	footerParts = append(footerParts, "Currently playing:")
	if m.debugInfo != "" {
		footerParts = append(footerParts, m.debugInfo)
	} else {
		footerParts = append(footerParts, "Nothing detected")
	}

	footerParts = append(footerParts, "")
	footerParts = append(footerParts, "Identified as:")
	if m.parsedArtist != "" && m.parsedTitle != "" {
		footerParts = append(footerParts, fmt.Sprintf("🎵 %s", m.parsedArtist))
		footerParts = append(footerParts, fmt.Sprintf("   %s", m.parsedTitle))
	} else {
		footerParts = append(footerParts, "Not parsed yet")
	}

	footer := helpStyle.Render(strings.Join(footerParts, "\n"))

	return lipgloss.NewStyle().
		Width(width).
		Height(m.height - 4).
		Render(leftContent + footer)
}

func (m Model) renderLyricsBox(width int) string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lavender).
		Width(width).
		Height(m.height - 4).
		Render(m.viewport.View())
}

func (m Model) renderProgressBar() string {
	if m.duration == 0 {
		return ""
	}

	barWidth := 30
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
