package ports

import "github.com/gabrielcapilla/yogo/internal/domain"

type YoutubeService interface {
	Search(query string) ([]domain.Song, error)
	GetStreamURL(songID string) (string, error)
}
