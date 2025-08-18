package ports

import "github.com/gabrielcapilla/yogo/internal/domain"

type StorageService interface {
	AddToHistory(entry domain.HistoryEntry) error
	GetHistory(limit int) ([]domain.HistoryEntry, error)
	Close() error
}
