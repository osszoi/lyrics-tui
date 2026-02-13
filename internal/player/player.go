package player

// Player provides an interface to interact with media players.
type Player interface {
	// CurrentSong returns the currently playing song's artist and title.
	CurrentSong() (artist, title string, err error)

	// Position returns the current playback position and total duration in seconds.
	Position() (position, duration float64, err error)
}
