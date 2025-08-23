package ui

import (
	"strings"
	"yogo/internal/domain"
	"yogo/internal/ports"

	tea "github.com/charmbracelet/bubbletea"
)

type searchItem struct{ song domain.Song }

func (i searchItem) FilterValue() string { return i.song.Title }
func (i searchItem) ID() string          { return i.song.ID }
func (i searchItem) ToSong() domain.Song { return i.song }

type youtubeDataSource struct {
	youtubeService ports.YoutubeService
	config         domain.Config
}

func (s youtubeDataSource) Fetch(query string) tea.Msg {
	songs, err := s.youtubeService.Search(query, s.config.SearchLimit)
	if err != nil {
		return ports.SearchErrorMsg{Err: err}
	}

	if len(songs) == 1 && strings.Contains(query, "http") {
		return ports.PlaySongMsg{Song: songs[0]}
	}

	return ports.SearchResultsMsg{Songs: songs}
}

func NewSearchModel(service ports.YoutubeService, cfg domain.Config, styles Styles) listAndFilterModel {
	return NewListAndFilterModel(
		"search",
		"Search for a song or paste a URL...",
		youtubeDataSource{youtubeService: service, config: cfg},
		styles,
	)
}
