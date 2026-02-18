package lyrics

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

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

func ParseVTT(content string) []Line {
	var lines []Line
	// matches HH:MM:SS.mmm or MM:SS.mmm
	tsRegex := regexp.MustCompile(`(\d+:)?(\d+):(\d+)\.(\d+)\s*-->`)

	rawLines := strings.Split(content, "\n")
	for i := 0; i < len(rawLines); i++ {
		if !tsRegex.MatchString(rawLines[i]) {
			continue
		}

		ts := parseVTTTimestamp(strings.TrimSpace(strings.Split(rawLines[i], "-->")[0]))

		var textParts []string
		for i++; i < len(rawLines); i++ {
			trimmed := strings.TrimSpace(rawLines[i])
			if trimmed == "" {
				break
			}
			// skip if it looks like another timestamp
			if tsRegex.MatchString(trimmed) {
				i--
				break
			}
			textParts = append(textParts, trimmed)
		}

		lines = append(lines, Line{
			Timestamp: ts,
			Text:      strings.Join(textParts, " "),
		})
	}

	return lines
}

func parseVTTTimestamp(s string) float64 {
	parts := strings.Split(s, ":")
	switch len(parts) {
	case 3:
		h, _ := strconv.ParseFloat(parts[0], 64)
		m, _ := strconv.ParseFloat(parts[1], 64)
		sec, _ := strconv.ParseFloat(parts[2], 64)
		return h*3600 + m*60 + sec
	case 2:
		m, _ := strconv.ParseFloat(parts[0], 64)
		sec, _ := strconv.ParseFloat(parts[1], 64)
		return m*60 + sec
	case 1:
		sec, _ := strconv.ParseFloat(parts[0], 64)
		return sec
	}
	return 0
}

func ExtractBetweenTags(text, tag string) string {
	openTag := "<" + tag + ">"
	closeTag := "</" + tag + ">"
	start := strings.Index(text, openTag)
	end := strings.Index(text, closeTag)
	if start == -1 || end == -1 || end <= start {
		return text
	}
	return strings.TrimSpace(text[start+len(openTag) : end])
}
