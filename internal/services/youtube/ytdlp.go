package youtube

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gabrielcapilla/yogo/internal/domain"
	"github.com/gabrielcapilla/yogo/internal/ports"
)

type ytdlpResponse struct {
	Entries []struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Uploader string `json:"uploader"`
		Channel  string `json:"channel"`
	} `json:"entries"`
}

type YTDLPClient struct{}

func NewYTDLPClient() ports.YoutubeService {
	return &YTDLPClient{}
}

func (c *YTDLPClient) Search(query string) ([]domain.Song, error) {
	searchQuery := fmt.Sprintf("ytsearch5:%s", query)

	cmd := exec.Command("yt-dlp", "--dump-single-json", searchQuery)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error al ejecutar yt-dlp search: %w", err)
	}

	var resp ytdlpResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		return nil, fmt.Errorf("error al parsear JSON de yt-dlp: %w", err)
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

	return songs, nil
}

func (c *YTDLPClient) GetStreamURL(songID string) (string, error) {
	cmd := exec.Command("yt-dlp", "-f", "bestaudio", "-g", songID)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error al ejecutar yt-dlp get-url: %w", err)
	}

	url := strings.TrimSpace(string(output))
	if url == "" {
		return "", fmt.Errorf("yt-dlp no devolvió ninguna URL para el ID: %s", songID)
	}

	firstURL := strings.Split(url, "\n")[0]

	return firstURL, nil
}
