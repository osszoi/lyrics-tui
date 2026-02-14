# Lyrics TUI

A terminal-based lyrics viewer with automatic song detection and synchronized lyric highlighting.

## What it does

Lyrics TUI displays song lyrics in your terminal with the following features:

- **Automatic song detection**: Detects currently playing songs via MPRIS (works with Spotify, VLC, and other MPRIS-compatible players on Linux)
- **Manual search**: Search for any song by typing the song name or artist
- **Synchronized lyrics**: Highlights the current line being sung in real-time
- **Timing adjustment**: Fine-tune lyric timing with +/- keys if synchronization is off
- **Intelligent parsing**: Uses Claude AI to parse natural language song queries
- **Caching**: Stores fetched lyrics locally to avoid repeated API calls

## Prerequisites

- Go 1.18 or higher
- Linux with D-Bus and MPRIS support
- Claude CLI tool installed and configured
- Genius API access token

## Installation

### Install via Go (recommended)

```bash
go install github.com/osszoi/lyrics-tui@latest
```

Make sure `$HOME/go/bin` is in your PATH.

### Download binary

Download the latest release from the [releases page](https://github.com/osszoi/lyrics-tui/releases) and extract it.

### Build from source

```bash
git clone <repository-url>
cd lyrics-tui
go build
```

## Configuration

The application works without configuration, but you can optionally set up a Genius API token for fallback lyrics when synced lyrics are unavailable.

Create a `.env` file:
```bash
GENIUS_ACCESS_TOKEN=your_genius_api_token_here
```

Get a token at https://genius.com/api-clients

## Usage

### Keyboard shortcuts

- **Tab**: Toggle auto-detect mode (automatically fetches lyrics for currently playing songs)
- **Enter**: Search for lyrics (when in manual mode)
- **+/-**: Adjust lyric timing offset by 0.1 seconds
- **/**: Toggle follow mode (auto-scroll to current line)
- **↑/↓ or j/k**: Scroll through lyrics manually
- **Esc**: Quit

### Modes

**Manual mode** (default): Type a song name or artist in the search box and press Enter.

**Auto-detect mode** (Tab): Automatically detects the currently playing song from your media player and fetches lyrics.

## How it works

1. Song queries are parsed using Claude AI to extract artist and title
2. The application first attempts to fetch time-synced lyrics from LRCLIB
3. If synced lyrics are unavailable, it falls back to plain lyrics from Genius
4. Lyrics are cached locally in the `songs/` directory
5. Playback position is retrieved via MPRIS to highlight the current line

## License

MIT
