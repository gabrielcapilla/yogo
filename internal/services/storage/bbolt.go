package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"time"
	"yogo/internal/domain"
	"yogo/internal/ports"

	"go.etcd.io/bbolt"
)

var historyBucket = []byte("history")

type BboltStore struct {
	db *bbolt.DB
}

func NewBboltStore(dbPath string) (ports.StorageService, error) {
	options := &bbolt.Options{Timeout: 1 * time.Second}
	db, err := bbolt.Open(dbPath, 0600, options)
	if err != nil {
		return nil, fmt.Errorf("could not open bbolt database: %w", err)
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(historyBucket)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("could not create history bucket: %w", err)
	}

	return &BboltStore{db: db}, nil
}

func (s *BboltStore) createHistoryKey(t time.Time, songID string) []byte {
	return []byte(fmt.Sprintf("%s:%s", t.UTC().Format(time.RFC3339Nano), songID))
}

func (s *BboltStore) findOldKeyForSong(b *bbolt.Bucket, songID string) (key, value []byte) {
	c := b.Cursor()
	suffix := []byte(":" + songID)

	for k, v := c.First(); k != nil; k, v = c.Next() {
		if bytes.HasSuffix(k, suffix) {
			return k, v
		}
	}
	return nil, nil
}

func (s *BboltStore) AddToHistory(entry domain.HistoryEntry) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(historyBucket)

		if oldKey, _ := s.findOldKeyForSong(b, entry.Song.ID); oldKey != nil {
			if err := b.Delete(oldKey); err != nil {
				return err
			}
		}

		entry.ResumeAt = 0
		entry.PlayedAt = time.Now()
		key := s.createHistoryKey(entry.PlayedAt, entry.Song.ID)

		value, err := json.Marshal(entry)
		if err != nil {
			return fmt.Errorf("error serializing history entry: %w", err)
		}

		return b.Put(key, value)
	})
}

func (s *BboltStore) UpdateHistoryEntryPosition(songID string, position int) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(historyBucket)

		oldKey, oldVal := s.findOldKeyForSong(b, songID)
		if oldKey == nil {
			return nil
		}

		var entry domain.HistoryEntry
		if err := json.Unmarshal(oldVal, &entry); err != nil {
			return err
		}

		if err := b.Delete(oldKey); err != nil {
			return err
		}

		entry.ResumeAt = position
		entry.PlayedAt = time.Now()
		newKey := s.createHistoryKey(entry.PlayedAt, songID)

		newValue, err := json.Marshal(entry)
		if err != nil {
			return err
		}

		return b.Put(newKey, newValue)
	})
}

func (s *BboltStore) GetHistory(limit int) ([]domain.HistoryEntry, error) {
	var entries []domain.HistoryEntry

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(historyBucket)

		return b.ForEach(func(k, v []byte) error {
			var entry domain.HistoryEntry
			if err := json.Unmarshal(v, &entry); err != nil {
				return nil
			}
			entries = append(entries, entry)
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].PlayedAt.After(entries[j].PlayedAt)
	})

	if len(entries) > limit {
		return entries[:limit], nil
	}

	return entries, nil
}

func (s *BboltStore) Close() error {
	return s.db.Close()
}
