package domain

import "time"

type Song struct {
	ID      string
	Title   string
	Artists []string
}

type HistoryEntry struct {
	Song     Song
	PlayedAt time.Time
	ResumeAt int
}
