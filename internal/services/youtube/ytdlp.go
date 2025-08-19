package youtube

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"yogo/internal/domain"
	"yogo/internal/logger"
	"yogo/internal/ports"
)

type ytdlpResponse struct {
	Entries []struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Uploader string `json:"uploader"`
		Channel  string `json:"channel"`
	} `json:"entries"`
}

type YTDLPClient struct {
	cookiesPath string
}

func NewYTDLPClient(cookiesPath string) ports.YoutubeService {
	if cookiesPath != "" {
		logger.Log.Printf("yt-dlp client initialized with cookies file: %s", cookiesPath)
	} else {
		logger.Log.Println("yt-dlp client initialized without cookies file.")
	}
	return &YTDLPClient{cookiesPath: cookiesPath}
}

func (c *YTDLPClient) executeYTDLP(args ...string) ([]byte, error) {
	if c.cookiesPath != "" {
		args = append([]string{"--cookies", c.cookiesPath}, args...)
	}

	cmd := exec.Command("yt-dlp", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if stderr.Len() > 0 {
		logger.Log.Printf("yt-dlp stderr output (args: %v): %s", args, stderr.String())
	}

	if err != nil {
		return nil, fmt.Errorf("yt-dlp failed with: %s", strings.TrimSpace(stderr.String()))
	}

	return stdout.Bytes(), nil
}

func (c *YTDLPClient) Search(query string) ([]domain.Song, error) {
	searchQuery := fmt.Sprintf("ytsearch5:%s", query)
	logger.Log.Printf("Executing yt-dlp search with query: '%s'", searchQuery)

	output, err := c.executeYTDLP("--dump-single-json", searchQuery)
	if err != nil {
		return nil, fmt.Errorf("yt-dlp search error: %w (details in log)", err)
	}

	var resp ytdlpResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		logger.Log.Printf("Error parsing yt-dlp JSON. Received JSON: %s", string(output))
		return nil, fmt.Errorf("error parsing yt-dlp JSON: %w", err)
	}

	var songs []domain.Song
	for _, entry := range resp.Entries {
		artist := entry.Uploader
		if artist == "" {
			artist = entry.Channel
		}
		songs = append(songs, domain.Song{
			ID:      entry.ID,
			Title:   entry.Title,
			Artists: []string{artist},
		})
	}
	logger.Log.Printf("Search successful, %d songs found.", len(songs))
	return songs, nil
}

func (c *YTDLPClient) GetStreamURL(songID string) (string, error) {
	logger.Log.Printf("Getting stream URL for ID: %s", songID)

	formatSelector := "bestaudio/best"
	output, err := c.executeYTDLP("-f", formatSelector, "-g", songID)
	if err != nil {
		return "", fmt.Errorf("yt-dlp get-url error: %w (details in log)", err)
	}

	url := strings.TrimSpace(string(output))
	if url == "" {
		return "", fmt.Errorf("yt-dlp returned no URL for ID %s", songID)
	}

	firstURL := strings.Split(url, "\n")[0]
	logger.Log.Printf("Successfully obtained stream URL for ID %s.", songID)
	return firstURL, nil
}
