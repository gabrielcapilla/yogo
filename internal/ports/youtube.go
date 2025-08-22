package ports

import (
	"yogo/internal/domain"
)

type YoutubeService interface {
	Search(query string, limit int) ([]domain.Song, error)
	GetStreamURL(songID string) (string, error)
}
