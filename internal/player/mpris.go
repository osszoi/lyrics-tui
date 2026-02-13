package player

import (
	"fmt"
	"os/exec"
	"strings"
)

// MPRISPlayer interacts with media players via the MPRIS D-Bus interface on Linux.
type MPRISPlayer struct{}

// NewMPRISPlayer creates a new MPRIS player interface.
func NewMPRISPlayer() *MPRISPlayer {
	return &MPRISPlayer{}
}

// CurrentSong retrieves the currently playing song via MPRIS.
func (p *MPRISPlayer) CurrentSong() (string, string, error) {
	cmd := exec.Command("bash", "-c", `
		player=$(busctl --user list | grep -oP 'org\.mpris\.MediaPlayer2\.\S+' | head -1)
		if [ -z "$player" ]; then
			exit 1
		fi

		metadata=$(busctl --user get-property "$player" /org/mpris/MediaPlayer2 org.mpris.MediaPlayer2.Player Metadata 2>/dev/null)

		artist=$(echo "$metadata" | grep -oP 'xesam:artist.*?as \d+ "\K[^"]+' | head -1)
		title=$(echo "$metadata" | grep -oP 'xesam:title.*?s "\K[^"]+' | head -1)

		echo "$artist"
		echo "$title"
	`)

	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("no media player found")
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		return "", "", fmt.Errorf("no media playing")
	}

	artist := strings.TrimSpace(lines[0])
	title := strings.TrimSpace(lines[1])

	if artist == "" || title == "" {
		return "", "", fmt.Errorf("incomplete metadata")
	}

	return artist, title, nil
}

// Position retrieves the current playback position and duration.
func (p *MPRISPlayer) Position() (float64, float64, error) {
	cmd := exec.Command("bash", "-c", `
		player=$(busctl --user list | grep -oP 'org\.mpris\.MediaPlayer2\.\S+' | head -1)
		if [ -z "$player" ]; then
			exit 1
		fi

		position=$(busctl --user get-property "$player" /org/mpris/MediaPlayer2 org.mpris.MediaPlayer2.Player Position 2>/dev/null | awk '{print $2}')

		metadata=$(busctl --user get-property "$player" /org/mpris/MediaPlayer2 org.mpris.MediaPlayer2.Player Metadata 2>/dev/null)
		duration=$(echo "$metadata" | grep -oP 'mpris:length.*?x \K\d+' | head -1)

		echo "$position"
		echo "$duration"
	`)

	output, err := cmd.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get position")
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	var positionMicroseconds float64
	var durationMicroseconds float64

	if len(lines) > 0 {
		fmt.Sscanf(lines[0], "%f", &positionMicroseconds)
	}
	if len(lines) > 1 {
		fmt.Sscanf(lines[1], "%f", &durationMicroseconds)
	}

	positionSeconds := positionMicroseconds / 1000000.0
	durationSeconds := durationMicroseconds / 1000000.0

	return positionSeconds, durationSeconds, nil
}
