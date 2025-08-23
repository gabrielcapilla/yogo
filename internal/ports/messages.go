package ports

import (
	"time"
	"yogo/internal/domain"
)

type FocusState int

const (
	GlobalFocus FocusState = iota
	ComponentFocus
)

type ChangeFocusMsg struct{ NewFocus FocusState }

type SearchResultsMsg struct{ Songs []domain.Song }
type SearchErrorMsg struct{ Err error }

type HistoryLoadedMsg struct{ Entries []domain.HistoryEntry }
type HistoryErrorMsg struct{ Err error }

type TickMsg time.Time
type PlaySongMsg struct{ Song domain.Song }
type StreamURLFetchedMsg struct {
	Song domain.Song
	URL  string
}
type SongNowPlayingMsg struct{ Song domain.Song }
type PlayErrorMsg struct{ Err error }
type PlayerStateUpdateMsg struct{ State PlayerState }
