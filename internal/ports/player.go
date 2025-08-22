package ports

type PlayerState struct {
	IsPlaying bool
	Position  float64
	Duration  float64
	Speed     float64
}

type PlayerService interface {
	Play(url string) error
	Pause() error
	Stop() error
	Seek(seconds int) error
	ChangeSpeed(delta float64) error
	ResetSpeed() error
	GetState() (PlayerState, error)
	Close() error
}
