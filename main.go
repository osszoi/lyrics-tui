package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"

	"lyrics-tui/internal/config"
	"lyrics-tui/internal/lyrics"
	"lyrics-tui/internal/parse"
	"lyrics-tui/internal/player"
	"lyrics-tui/internal/ui"
)

func main() {
	godotenv.Load()

	cfg := config.Load()

	geniusToken := os.Getenv("GENIUS_ACCESS_TOKEN")

	lrclibProvider := lyrics.NewLRCLIBProvider()
	geniusProvider := lyrics.NewGeniusProvider(geniusToken)
	cache := lyrics.NewCache("songs")
	lyricsService := lyrics.NewService(lrclibProvider, geniusProvider, cache)

	mprisPlayer := player.NewMPRISPlayer()

	parser, err := parse.NewProviderFromConfig(cfg)
	if err != nil {
		fmt.Printf("Error creating parser: %v\n", err)
		os.Exit(1)
	}

	model := ui.NewModel(lyricsService, mprisPlayer, parser, cfg)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
