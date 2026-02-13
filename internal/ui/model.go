package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"lyrics-tui/internal/lyrics"
	"lyrics-tui/internal/player"
)

// Parser defines the interface for parsing song queries.
type Parser interface {
	Parse(query string) (artist, title string, err error)
}

// Model is the main Bubble Tea model for the lyrics TUI.
type Model struct {
	// Dependencies
	lyricsService *lyrics.Service
	player        player.Player
	parser        Parser

	// UI components
	input    textinput.Model
	viewport viewport.Model

	// Current song state
	artist          string
	title           string
	lyrics          string
	syncedLyrics    []lyrics.Line
	hasSyncedLyrics bool

	// Playback state
	playbackPosition    float64
	duration            float64
	offset              float64
	ignorePositionUntil time.Time

	// UI state
	followMode     bool
	autoDetectMode bool
	searching      bool
	ready          bool
	width          int
	height         int

	// MPRIS tracking
	lastDetectedSong string
	mprisArtist      string
	mprisTitle       string

	// Parsing state
	parsedArtist string
	parsedTitle  string

	// Debug info
	debugInfo string
	err       error
}

// NewModel creates a new TUI model with the given dependencies.
func NewModel(lyricsService *lyrics.Service, player player.Player, parser Parser) Model {
	ti := textinput.New()
	ti.Placeholder = "Type song name..."
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 80

	vp := viewport.New(80, 20)
	vp.SetContent("🎤 Waiting for you to drop some song names...\n\n" +
		"I'm just sitting here, twiddling my bytes.\n" +
		"Type something already! 😴")

	return Model{
		lyricsService: lyricsService,
		player:        player,
		parser:        parser,
		input:         ti,
		viewport:      vp,
		followMode:    true,
		ready:         false,
	}
}

// Init initializes the model and returns initial commands.
func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, tickEverySecond(), tickPosition())
}
