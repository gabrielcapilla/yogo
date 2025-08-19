package storage

import (
	"path/filepath"
	"testing"
	"time"
	"yogo/internal/domain"

	"github.com/stretchr/testify/require"
)

func TestBboltStore_History(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	store, err := NewBboltStore(dbPath)
	require.NoError(t, err, "Failed to create new bbolt store")
	defer store.Close()

	entry1 := domain.HistoryEntry{Song: domain.Song{ID: "song1_id", Title: "Song 1"}}
	store.AddToHistory(entry1)
	time.Sleep(2 * time.Millisecond)

	entry2 := domain.HistoryEntry{Song: domain.Song{ID: "song2_id", Title: "Song 2"}}
	store.AddToHistory(entry2)
	time.Sleep(2 * time.Millisecond)

	entry3 := domain.HistoryEntry{Song: domain.Song{ID: "song3_id", Title: "Song 3"}}
	store.AddToHistory(entry3)
	time.Sleep(2 * time.Millisecond)

	history, err := store.GetHistory(10)

	require.NoError(t, err, "GetHistory should not return an error")
	require.Len(t, history, 3, "History should contain 3 entries")
	require.Equal(t, "song3_id", history[0].Song.ID, "The most recently added song should be first")
	require.Equal(t, "song2_id", history[1].Song.ID)
	require.Equal(t, "song1_id", history[2].Song.ID)

	duplicateEntry1 := domain.HistoryEntry{Song: domain.Song{ID: "song1_id", Title: "Song 1 Updated"}}

	store.AddToHistory(duplicateEntry1)

	historyAfterUpdate, err := store.GetHistory(10)

	require.NoError(t, err)
	require.Len(t, historyAfterUpdate, 3, "History should still contain only 3 unique entries")
	require.Equal(t, "song1_id", historyAfterUpdate[0].Song.ID, "The updated song should now be first")
	require.Equal(t, "song3_id", historyAfterUpdate[1].Song.ID)
	require.Equal(t, "song2_id", historyAfterUpdate[2].Song.ID)

	limitedHistory, err := store.GetHistory(2)

	require.NoError(t, err)
	require.Len(t, limitedHistory, 2, "History should be truncated to the limit")
	require.Equal(t, "song1_id", limitedHistory[0].Song.ID)
	require.Equal(t, "song3_id", limitedHistory[1].Song.ID)
}
