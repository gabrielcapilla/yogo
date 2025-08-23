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
	err = store.AddToHistory(entry1)
	require.NoError(t, err)
	time.Sleep(2 * time.Millisecond)

	entry2 := domain.HistoryEntry{Song: domain.Song{ID: "song2_id", Title: "Song 2"}}
	err = store.AddToHistory(entry2)
	require.NoError(t, err)
	time.Sleep(2 * time.Millisecond)

	entry3 := domain.HistoryEntry{Song: domain.Song{ID: "song3_id", Title: "Song 3"}}
	err = store.AddToHistory(entry3)
	require.NoError(t, err)
	time.Sleep(2 * time.Millisecond)

	history, err := store.GetHistory(10)

	require.NoError(t, err, "GetHistory should not return an error")
	require.Len(t, history, 3, "History should contain 3 entries")
	require.Equal(t, "song3_id", history[0].Song.ID, "The most recently added song should be first")
	require.Equal(t, "song2_id", history[1].Song.ID)
	require.Equal(t, "song1_id", history[2].Song.ID)

	duplicateEntry1 := domain.HistoryEntry{Song: domain.Song{ID: "song1_id", Title: "Song 1 Updated"}}

	err = store.AddToHistory(duplicateEntry1)
	require.NoError(t, err)

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

	err = store.UpdateHistoryEntryPosition("song2_id", 120)
	require.NoError(t, err)

	historyAfterPositionUpdate, err := store.GetHistory(10)
	require.NoError(t, err)
	require.Len(t, historyAfterPositionUpdate, 3)
	require.Equal(t, "song2_id", historyAfterPositionUpdate[0].Song.ID, "Song 2 should be first after position update")
	require.Equal(t, 120, historyAfterPositionUpdate[0].ResumeAt, "ResumeAt should be updated")
	require.Equal(t, "song1_id", historyAfterPositionUpdate[1].Song.ID)
}
