package ui

import (
	"yogo/internal/domain"
	"yogo/internal/ports"
)

type searchResultsMsg struct {
	songs []domain.Song
}

type searchErrorMsg struct {
	err error
}

type playSongMsg struct {
	song domain.Song
}

type streamURLFetchedMsg struct {
	song domain.Song
	url  string
}

type songNowPlayingMsg struct {
	song domain.Song
}

type playErrorMsg struct {
	err error
}

type HistoryLoadedMsg struct {
	Entries []domain.HistoryEntry
}

type HistoryErrorMsg struct {
	Err error
}

type playerStateUpdateMsg struct {
	state ports.PlayerState
}

type changeFocusMsg struct {
	newFocus focusState
}
