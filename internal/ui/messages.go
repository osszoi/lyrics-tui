package ui

import (
	"time"

	"lyrics-tui/internal/lyrics"
)

// tickMsg triggers periodic MPRIS polling.
type tickMsg time.Time

// positionTickMsg triggers playback position updates.
type positionTickMsg time.Time

// mprisData contains currently playing song metadata from MPRIS.
type mprisData struct {
	artist string
	title  string
	err    error
}

// playbackPosition contains current playback state.
type playbackPosition struct {
	position float64
	duration float64
	err      error
}

// parsedResult contains the result of parsing a song query.
type parsedResult struct {
	artist      string
	title       string
	mprisArtist string
	mprisTitle  string
	err         error
}

// searchResult contains fetched lyrics.
type searchResult struct {
	song        *lyrics.Song
	mprisArtist string
	mprisTitle  string
	err         error
}

type aiLyricsResult struct {
	artist      string
	title       string
	lyrics      string
	query       string
	mprisArtist string
	mprisTitle  string
	err         error
}
