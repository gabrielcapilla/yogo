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

var execCommand = exec.Command

type ytdlpSearchResponse struct {
	Entries []struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Uploader string `json:"uploader"`
		Channel  string `json:"channel"`
	} `json:"entries"`
}

type ytdlpGenericSearchResponse struct {
	Entries []struct {
		Type  string `json:"_type"`
		ID    string `json:"id"`
		Title string `json:"title"`
	} `json:"entries"`
}

type ytdlpPlaylistResponse struct {
	Entries []struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Uploader string `json:"uploader"`
	} `json:"entries"`
}

type YTDLPClient struct {
	cookiesPath string
}

func NewYTDLPClient(cookiesPath string) ports.YoutubeService {
	if cookiesPath != "" {
		logger.Log.Info().Str("cookies_path", cookiesPath).Msg("yt-dlp client initialized with cookies file")
	} else {
		logger.Log.Info().Msg("yt-dlp client initialized without cookies file")
	}
	return &YTDLPClient{cookiesPath: cookiesPath}
}

func (c *YTDLPClient) executeYTDLP(args ...string) ([]byte, error) {
	if c.cookiesPath != "" {
		args = append([]string{"--cookies", c.cookiesPath}, args...)
	}

	cmd := execCommand("yt-dlp", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if stderr.Len() > 0 {
		logger.Log.Warn().Strs("args", args).Str("stderr", stderr.String()).Msg("yt-dlp stderr output")
	}

	if err != nil {
		return nil, fmt.Errorf("yt-dlp failed with: %s", strings.TrimSpace(stderr.String()))
	}

	return stdout.Bytes(), nil
}

func (c *YTDLPClient) Search(query string) ([]domain.Song, error) {
	searchQuery := fmt.Sprintf("ytsearch5:%s", query)
	logger.Log.Info().Str("query", searchQuery).Msg("Executing yt-dlp search")

	output, err := c.executeYTDLP("--dump-single-json", "--", searchQuery)
	if err != nil {
		return nil, fmt.Errorf("yt-dlp search error: %w (details in log)", err)
	}

	var resp ytdlpSearchResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		logger.Log.Error().Err(err).Str("json_output", string(output)).Msg("Error parsing yt-dlp JSON")
		return nil, fmt.Errorf("error parsing yt-dlp JSON: %w", err)
	}

	songs := make([]domain.Song, 0)
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
	logger.Log.Info().Int("song_count", len(songs)).Msg("Search successful")
	return songs, nil
}

func (c *YTDLPClient) GetStreamURL(songID string) (string, error) {
	logger.Log.Info().Str("song_id", songID).Msg("Getting stream URL")

	formatSelector := "bestaudio/best"
	output, err := c.executeYTDLP("-f", formatSelector, "-g", "--", songID)
	if err != nil {
		return "", fmt.Errorf("yt-dlp get-url error: %w (details in log)", err)
	}

	url := strings.TrimSpace(string(output))
	if url == "" {
		return "", fmt.Errorf("yt-dlp returned no URL for ID %s", songID)
	}

	firstURL := strings.Split(url, "\n")[0]
	logger.Log.Info().Str("song_id", songID).Msg("Successfully obtained stream URL")
	return firstURL, nil
}

func (c *YTDLPClient) SearchPlaylists(query string) ([]domain.Playlist, error) {
	searchQuery := fmt.Sprintf("ytsearch10:%s", query)
	logger.Log.Info().Str("query", searchQuery).Msg("Executing yt-dlp playlist search")

	output, err := c.executeYTDLP("--dump-single-json", "--", searchQuery)
	if err != nil {
		return nil, fmt.Errorf("yt-dlp playlist search error: %w", err)
	}

	var resp ytdlpGenericSearchResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		logger.Log.Error().Err(err).Str("json_output", string(output)).Msg("Error parsing yt-dlp JSON for playlists")
		return nil, fmt.Errorf("error parsing yt-dlp JSON for playlists: %w", err)
	}

	playlists := make([]domain.Playlist, 0)
	for _, entry := range resp.Entries {
		if entry.Type == "playlist" {
			playlists = append(playlists, domain.Playlist{
				ID:    entry.ID,
				Title: entry.Title,
			})
		}
	}
	logger.Log.Info().Int("playlist_count", len(playlists)).Msg("Playlist search successful")
	return playlists, nil
}

func (c *YTDLPClient) GetPlaylistSongs(playlistID string) ([]domain.Song, error) {
	logger.Log.Info().Str("playlist_id", playlistID).Msg("Getting playlist songs")
	playlistURL := fmt.Sprintf("https://www.youtube.com/playlist?list=%s", playlistID)

	output, err := c.executeYTDLP("--flat-playlist", "--dump-single-json", "--", playlistURL)
	if err != nil {
		return nil, fmt.Errorf("yt-dlp get-playlist-songs error: %w", err)
	}

	var resp ytdlpPlaylistResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		logger.Log.Error().Err(err).Str("json_output", string(output)).Msg("Error parsing yt-dlp playlist songs JSON")
		return nil, fmt.Errorf("error parsing yt-dlp playlist songs JSON: %w", err)
	}

	songs := make([]domain.Song, 0)
	for _, entry := range resp.Entries {
		songs = append(songs, domain.Song{
			ID:      entry.ID,
			Title:   entry.Title,
			Artists: []string{entry.Uploader},
		})
	}
	logger.Log.Info().Int("song_count", len(songs)).Str("playlist_id", playlistID).Msg("Successfully fetched playlist songs")
	return songs, nil
}
