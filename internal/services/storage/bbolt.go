package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
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

func (s *BboltStore) findAndDeleteOldEntry(b *bbolt.Bucket, songID string) error {
	c := b.Cursor()
	prefix := []byte(":")
	idBytes := []byte(songID)

	for k, _ := c.First(); k != nil; k, _ = c.Next() {
		parts := bytes.Split(k, prefix)
		if len(parts) == 2 && bytes.Equal(parts[1], idBytes) {
			return c.Delete()
		}
	}
	return nil
}

func (s *BboltStore) AddToHistory(entry domain.HistoryEntry) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(historyBucket)

		if err := s.findAndDeleteOldEntry(b, entry.Song.ID); err != nil {
			return err
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

		var oldEntry domain.HistoryEntry
		var oldKey []byte

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if strings.HasSuffix(string(k), songID) {
				if err := json.Unmarshal(v, &oldEntry); err != nil {
					return err
				}
				oldKey = k
				break
			}
		}

		if oldKey == nil {
			return nil
		}

		if err := b.Delete(oldKey); err != nil {
			return err
		}

		oldEntry.ResumeAt = position
		oldEntry.PlayedAt = time.Now()
		newKey := s.createHistoryKey(oldEntry.PlayedAt, songID)

		newValue, err := json.Marshal(oldEntry)
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
		c := b.Cursor()

		for k, v := c.Last(); k != nil && len(entries) < limit; k, v = c.Prev() {
			var entry domain.HistoryEntry
			if err := json.Unmarshal(v, &entry); err != nil {
				return fmt.Errorf("error deserializing history entry: %w", err)
			}
			entries = append(entries, entry)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return entries, nil
}

func (s *BboltStore) Close() error {
	return s.db.Close()
}
