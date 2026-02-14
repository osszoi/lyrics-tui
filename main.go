package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"

	"lyrics-tui/internal/lyrics"
	"lyrics-tui/internal/parse"
	"lyrics-tui/internal/player"
	"lyrics-tui/internal/ui"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env file")
		//os.Exit(1)
	}

	geniusToken := os.Getenv("GENIUS_ACCESS_TOKEN")
	if geniusToken == "" {
		fmt.Println("GENIUS_ACCESS_TOKEN not set in .env file")
		//os.Exit(1)
		geniusToken = ""
	}

	lrclibProvider := lyrics.NewLRCLIBProvider()
	geniusProvider := lyrics.NewGeniusProvider(geniusToken)
	cache := lyrics.NewCache("songs")
	lyricsService := lyrics.NewService(lrclibProvider, geniusProvider, cache)

	mprisPlayer := player.NewMPRISPlayer()

	claudeParser := parse.NewClaudeParser()

	model := ui.NewModel(lyricsService, mprisPlayer, claudeParser)

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
