package domain

type Config struct {
	CookiesPath  string `mapstructure:"cookiesPath"`
	HistoryLimit int    `mapstructure:"historyLimit"`
}
