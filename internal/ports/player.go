package ports

type PlayerService interface {
	Play(url string) error
	Pause() error
	Stop() error
}
