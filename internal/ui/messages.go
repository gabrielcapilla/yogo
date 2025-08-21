package ui

import (
	"time"
	"yogo/internal/domain"
	"yogo/internal/ports"
)

type focusState int

const (
	globalFocus focusState = iota
	searchFocus
)

type changeFocusMsg struct{ newFocus focusState }

type searchResultsMsg struct{ songs []domain.Song }
type searchErrorMsg struct{ err error }

type tickMsg time.Time
type playSongMsg struct{ song domain.Song }
type streamURLFetchedMsg struct {
	song domain.Song
	url  string
}
type songNowPlayingMsg struct{ song domain.Song }
type playErrorMsg struct{ err error }
type playerStateUpdateMsg struct{ state ports.PlayerState }
