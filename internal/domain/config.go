package domain

type PlaybackConfig struct {
	Loop               bool `mapstructure:"loop"`
	SavePositionOnQuit bool `mapstructure:"savePositionOnQuit"`
}

type Config struct {
	CookiesPath  string         `mapstructure:"cookiesPath"`
	HistoryLimit int            `mapstructure:"historyLimit"`
	SearchLimit  int            `mapstructure:"searchLimit"`
	Playback     PlaybackConfig `mapstructure:"playback"`
}
