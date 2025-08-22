package youtube

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"yogo/internal/domain"
	"yogo/internal/logger"
	"yogo/internal/ports"

	"github.com/buger/jsonparser"
)

var execCommand = exec.Command

type YoutubeClient struct {
	cookiesPath string
	httpClient  *http.Client
}

func NewYoutubeClient(cookiesPath string) ports.YoutubeService {
	return &YoutubeClient{
		cookiesPath: cookiesPath,
		httpClient:  &http.Client{},
	}
}

func (c *YoutubeClient) Search(query string, limit int) ([]domain.Song, error) {
	if strings.HasPrefix(query, "http") {
		return c.getSongInfoFromURL(query)
	}

	return c.scrapeSearchResults(query, limit)
}

func (c *YoutubeClient) scrapeSearchResults(query string, limit int) ([]domain.Song, error) {
	searchURL := "https://www.youtube.com/results?search_query=" + url.QueryEscape(query)

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept-Language", "en")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("youtube returned non-200 status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	htmlBody := string(body)
	splittedScript := strings.Split(htmlBody, `var ytInitialData = `)
	if len(splittedScript) < 2 {
		splittedScript = strings.Split(htmlBody, `window["ytInitialData"] = `)
	}
	if len(splittedScript) < 2 {
		return nil, errors.New("could not find ytInitialData script block")
	}
	jsonBlock := strings.Split(splittedScript[1], `;</script>`)[0]
	jsonData := []byte(jsonBlock)

	contentRoot, _, _, err := jsonparser.Get(jsonData, "contents", "twoColumnSearchResultsRenderer", "primaryContents", "sectionListRenderer", "contents")
	if err != nil {
		return nil, fmt.Errorf("could not parse initial JSON structure: %w", err)
	}

	var resultsContents []byte
	jsonparser.ArrayEach(contentRoot, func(contentBlock []byte, dataType jsonparser.ValueType, offset int, err error) {
		contents, _, _, err := jsonparser.Get(contentBlock, "itemSectionRenderer", "contents")
		if err == nil {
			resultsContents = contents
		}
	})

	if resultsContents == nil {
		return nil, errors.New("could not find video results content block in JSON")
	}

	var songs []domain.Song
	count := 0
	jsonparser.ArrayEach(resultsContents, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if count >= limit {
			return
		}
		videoID, err := jsonparser.GetString(value, "videoRenderer", "videoId")
		if err != nil {
			return
		}

		title, _ := jsonparser.GetString(value, "videoRenderer", "title", "runs", "[0]", "text")
		uploader, _ := jsonparser.GetString(value, "videoRenderer", "ownerText", "runs", "[0]", "text")

		songs = append(songs, domain.Song{
			ID:      videoID,
			Title:   title,
			Artists: []string{uploader},
		})
		count++
	})

	return songs, nil
}

func (c *YoutubeClient) getSongInfoFromURL(url string) ([]domain.Song, error) {
	output, err := c.executeYTDLP("--dump-single-json", "--", url)
	if err != nil {
		return nil, err
	}

	var entry struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Uploader string `json:"uploader"`
		Channel  string `json:"channel"`
	}
	if err := json.Unmarshal(output, &entry); err != nil {
		return nil, err
	}

	artist := entry.Uploader
	if artist == "" {
		artist = entry.Channel
	}
	song := domain.Song{
		ID:      entry.ID,
		Title:   entry.Title,
		Artists: []string{artist},
	}
	return []domain.Song{song}, nil
}

func (c *YoutubeClient) GetStreamURL(songID string) (string, error) {
	logger.Log.Info().Str("song_id", songID).Msg("Getting stream URL via yt-dlp")

	args := []string{"-f", "bestaudio/best", "-g", "--", songID}
	output, err := c.executeYTDLP(args...)
	if err != nil {
		return "", err
	}

	url := strings.TrimSpace(string(output))
	if url == "" {
		return "", fmt.Errorf("yt-dlp returned no URL for ID %s", songID)
	}

	return strings.Split(url, "\n")[0], nil
}

func (c *YoutubeClient) executeYTDLP(args ...string) ([]byte, error) {
	if c.cookiesPath != "" {
		fullArgs := append([]string{"--cookies", c.cookiesPath}, args...)
		args = fullArgs
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
