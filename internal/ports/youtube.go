package ports

import (
	"yogo/internal/domain"
)

type YoutubeService interface {
	Search(query string, limit int) ([]domain.Song, error)
	GetSongInfo(url string) (domain.Song, error)
}
