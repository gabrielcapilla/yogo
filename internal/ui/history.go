package ui

import (
	"yogo/internal/domain"
	"yogo/internal/ports"

	tea "github.com/charmbracelet/bubbletea"
)

type historyItem struct{ entry domain.HistoryEntry }

func (i historyItem) FilterValue() string { return i.entry.Song.Title }
func (i historyItem) ID() string          { return i.entry.Song.ID }
func (i historyItem) ToSong() domain.Song { return i.entry.Song }
func (i historyItem) ResumeAt() int       { return i.entry.ResumeAt }

type historyDataSource struct {
	storageService ports.StorageService
	config         domain.Config
}

func (s historyDataSource) Fetch(query string) tea.Msg {
	entries, err := s.storageService.GetHistory(s.config.HistoryLimit)
	if err != nil {
		return ports.HistoryErrorMsg{Err: err}
	}
	return ports.HistoryLoadedMsg{Entries: entries}
}

func NewHistoryModel(service ports.StorageService, cfg domain.Config, styles Styles) listAndFilterModel {
	return NewListAndFilterModel(
		"history",
		"Filter history...",
		historyDataSource{storageService: service, config: cfg},
		styles,
	)
}
