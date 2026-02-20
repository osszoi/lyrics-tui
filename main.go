package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"

	"lyrics-tui/internal/config"
	"lyrics-tui/internal/lyrics"
	"lyrics-tui/internal/parse"
	"lyrics-tui/internal/player"
	"lyrics-tui/internal/ui"
)

var Version = "dev"

func main() {
	godotenv.Load()

	cfg := config.Load()

	geniusToken := os.Getenv("GENIUS_ACCESS_TOKEN")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting home directory: %v\n", err)
		os.Exit(1)
	}
	cacheDir := filepath.Join(homeDir, ".config", "lyrics", "cached_songs")

	lrclibProvider := lyrics.NewLRCLIBProvider()
	geniusProvider := lyrics.NewGeniusProvider(geniusToken)
	cache := lyrics.NewCache(cacheDir)
	lyricsService := lyrics.NewService(lrclibProvider, geniusProvider, cache)

	mprisPlayer := player.NewMPRISPlayer()

	parser, err := parse.NewProviderFromConfig(cfg)
	if err != nil {
		fmt.Printf("Error creating parser: %v\n", err)
		os.Exit(1)
	}

	model := ui.NewModel(lyricsService, mprisPlayer, parser, cfg, Version)

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
