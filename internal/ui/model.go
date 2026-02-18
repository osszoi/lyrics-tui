package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"lyrics-tui/internal/config"
	"lyrics-tui/internal/lyrics"
	"lyrics-tui/internal/parse"
	"lyrics-tui/internal/player"
)

type Model struct {
	lyricsService *lyrics.Service
	player        player.Player
	parser        parse.Provider
	config        *config.Config
	version       string

	input    textinput.Model
	viewport viewport.Model

	artist          string
	title           string
	lyrics          string
	syncedLyrics    []lyrics.Line
	hasSyncedLyrics bool

	playbackPosition    float64
	duration            float64
	offset              float64
	ignorePositionUntil time.Time

	followMode     bool
	autoDetectMode bool
	searching      bool
	ready          bool
	width          int
	height         int

	lastDetectedSong string
	mprisArtist      string
	mprisTitle       string

	parsedArtist string
	parsedTitle  string

	debugInfo string
	err       error

	// settings modal
	settingsOpen        bool
	settingsCursor      int
	settingsProviderIdx int
	settingsModel       textinput.Model
	settingsAPIKey      textinput.Model

	// search modal
	searchModalOpen bool

	// cached songs modal
	cachedSongsModalOpen bool
	cachedSongs          []lyrics.CachedSongEntry
	cachedSongsFiltered  []lyrics.CachedSongEntry
	cachedSongsCursor    int
	cachedSongsFilter    textinput.Model
}

func NewModel(lyricsService *lyrics.Service, player player.Player, parser parse.Provider, cfg *config.Config, version string) Model {
	ti := textinput.New()
	ti.Placeholder = "Type song name..."
	ti.CharLimit = 200
	ti.Width = 40

	vp := viewport.New(80, 20)
	vp.SetContent("Waiting for you to drop some song names...\n\n" +
		"I'm just sitting here, twiddling my bytes.\n" +
		"Type something already!")

	sm := textinput.New()
	sm.Placeholder = "model name"
	sm.CharLimit = 100
	sm.Width = 30

	sa := textinput.New()
	sa.Placeholder = "leave empty for env var"
	sa.CharLimit = 200
	sa.Width = 30

	cf := textinput.New()
	cf.Placeholder = "filter..."
	cf.CharLimit = 100
	cf.Width = 40

	return Model{
		lyricsService:     lyricsService,
		player:            player,
		parser:            parser,
		config:            cfg,
		version:           version,
		input:             ti,
		viewport:          vp,
		followMode:        true,
		settingsModel:     sm,
		settingsAPIKey:    sa,
		cachedSongsFilter: cf,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, tickEverySecond(), tickPosition())
}
