package ports

import "yogo/internal/domain"

type YoutubeService interface {
	Search(query string) ([]domain.Song, error)
	GetStreamURL(songID string) (string, error)
	SearchPlaylists(query string) ([]domain.Playlist, error)
	GetPlaylistSongs(playlistID string) ([]domain.Song, error)
}
