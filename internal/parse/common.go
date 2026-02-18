package parse

type parsedSong struct {
	Artist string `json:"artist"`
	Title  string `json:"title"`
}

type lyricsResult struct {
	Artist string `json:"artist"`
	Song   string `json:"song"`
	Lyrics string `json:"lyrics"`
}
