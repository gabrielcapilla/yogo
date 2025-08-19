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

type YTDLPClient struct{}

func NewYTDLPClient() ports.YoutubeService {
	return &YTDLPClient{}
}

func executeYTDLP(args ...string) ([]byte, error) {
	cmd := exec.Command("yt-dlp", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if stderr.Len() > 0 {
		logger.Log.Printf("Salida de stderr de yt-dlp (args: %v): %s", args, stderr.String())
	}

	if err != nil {
		return nil, fmt.Errorf("yt-dlp falló con: %s", strings.TrimSpace(stderr.String()))
	}

	return stdout.Bytes(), nil
}

func (c *YTDLPClient) Search(query string) ([]domain.Song, error) {
	searchQuery := fmt.Sprintf("ytsearch5:%s", query)
	logger.Log.Printf("Ejecutando búsqueda en yt-dlp con query: '%s'", searchQuery)

	output, err := executeYTDLP("--dump-single-json", searchQuery)
	if err != nil {
		return nil, fmt.Errorf("error en yt-dlp search: %w (detalles en el log)", err)
	}

	var resp ytdlpResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		logger.Log.Printf("Error al parsear JSON de yt-dlp. JSON recibido: %s", string(output))
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
	logger.Log.Printf("Búsqueda exitosa, %d canciones encontradas.", len(songs))
	return songs, nil
}

func (c *YTDLPClient) GetStreamURL(songID string) (string, error) {
	logger.Log.Printf("Obteniendo URL de stream para el ID: %s", songID)

	output, err := executeYTDLP("-f", "bestaudio", "-g", songID)
	if err != nil {
		return "", fmt.Errorf("error en yt-dlp get-url: %w (detalles en el log)", err)
	}

	url := strings.TrimSpace(string(output))
	if url == "" {
		return "", fmt.Errorf("yt-dlp no devolvió ninguna URL para el ID %s", songID)
	}

	firstURL := strings.Split(url, "\n")[0]
	logger.Log.Printf("URL de stream obtenida con éxito para ID %s.", songID)
	return firstURL, nil
}
