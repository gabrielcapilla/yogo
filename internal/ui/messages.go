package ui

import (
	"yogo/internal/domain"
)

type searchResultsMsg struct {
	songs []domain.Song
}

type searchErrorMsg struct {
	err error
}
