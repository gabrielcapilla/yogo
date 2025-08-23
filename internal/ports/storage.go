package ports

import "yogo/internal/domain"

type StorageService interface {
	AddToHistory(entry domain.HistoryEntry) error
	GetHistory(limit int) ([]domain.HistoryEntry, error)
	UpdateHistoryEntryPosition(songID string, position int) error
	DeleteFromHistory(songID string) error
	Close() error
}
