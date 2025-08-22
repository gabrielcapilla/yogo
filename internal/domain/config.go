package domain

type Config struct {
	CookiesPath  string `mapstructure:"cookiesPath"`
	HistoryLimit int    `mapstructure:"historyLimit"`
	SearchLimit  int    `mapstructure:"searchLimit"`
}
