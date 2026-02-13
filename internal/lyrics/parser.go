package lyrics

import (
	"fmt"
	"regexp"
	"strings"
)

// ParseLRC parses LRC format lyrics into a slice of timestamped lines.
func ParseLRC(lrcContent string) []Line {
	var lines []Line
	re := regexp.MustCompile(`\[(\d+):(\d+)\.(\d+)\](.*)`)

	for _, line := range strings.Split(lrcContent, "\n") {
		matches := re.FindStringSubmatch(line)
		if len(matches) == 5 {
			minutes := 0
			seconds := 0
			centiseconds := 0
			fmt.Sscanf(matches[1], "%d", &minutes)
			fmt.Sscanf(matches[2], "%d", &seconds)
			fmt.Sscanf(matches[3], "%d", &centiseconds)

			timestamp := float64(minutes*60+seconds) + float64(centiseconds)/100.0
			text := strings.TrimSpace(matches[4])

			if text != "" {
				lines = append(lines, Line{
					Timestamp: timestamp,
					Text:      text,
				})
			}
		}
	}

	return lines
}
