package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"lyrics-tui/internal/lyrics"
)

func tickEverySecond() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func tickPosition() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return positionTickMsg(t)
	})
}

func (m Model) detectCurrentSong() tea.Cmd {
	return func() tea.Msg {
		artist, title, err := m.player.CurrentSong()
		if err != nil {
			return mprisData{err: err}
		}

		return mprisData{
			artist: artist,
			title:  title,
		}
	}
}

func (m Model) getPlaybackPosition() tea.Cmd {
	return func() tea.Msg {
		position, duration, err := m.player.Position()
		if err != nil {
			return playbackPosition{err: err}
		}

		return playbackPosition{
			position: position,
			duration: duration,
		}
	}
}

func (m Model) searchLyrics(query string) tea.Cmd {
	if m.config.AILyrics {
		return m.fetchAILyrics(query, "", "")
	}
	return m.searchLyricsWithMpris(query, "", "")
}

func (m Model) searchLyricsWithMpris(query, mprisArtist, mprisTitle string) tea.Cmd {
	return func() tea.Msg {
		artist, title, err := m.parser.Parse(query)
		if err != nil {
			return parsedResult{
				err:         fmt.Errorf("failed to parse: %w", err),
				mprisArtist: mprisArtist,
				mprisTitle:  mprisTitle,
			}
		}

		return parsedResult{
			artist:      artist,
			title:       title,
			mprisArtist: mprisArtist,
			mprisTitle:  mprisTitle,
		}
	}
}

func (m Model) fetchLyrics(artist, title, mprisArtist, mprisTitle string) tea.Cmd {
	return func() tea.Msg {
		song, err := m.lyricsService.Fetch(artist, title)
		if err != nil {
			return searchResult{
				err:         err,
				mprisArtist: mprisArtist,
				mprisTitle:  mprisTitle,
			}
		}

		return searchResult{
			song:        song,
			mprisArtist: mprisArtist,
			mprisTitle:  mprisTitle,
		}
	}
}

func (m Model) fetchAILyrics(query, mprisArtist, mprisTitle string) tea.Cmd {
	return func() tea.Msg {
		prompt := fmt.Sprintf(`Give me the lyrics for "%s". Surround them by <lyrics> tag and add VTT`, query)
		response, err := m.parser.Generate(prompt)
		if err != nil {
			return aiLyricsResult{err: err, query: query, mprisArtist: mprisArtist, mprisTitle: mprisTitle}
		}

		extracted := lyrics.ExtractBetweenTags(response, "lyrics")
		syncedLines := lyrics.ParseVTT(extracted)

		return aiLyricsResult{
			lyricsText:  extracted,
			syncedLines: syncedLines,
			query:       query,
			mprisArtist: mprisArtist,
			mprisTitle:  mprisTitle,
		}
	}
}
