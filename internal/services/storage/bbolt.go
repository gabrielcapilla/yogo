package storage

import (
	"encoding/json"
	"fmt"

	"yogo/internal/domain"
	"yogo/internal/ports"

	"go.etcd.io/bbolt"
)

var historyBucket = []byte("history")

type BboltStore struct {
	db *bbolt.DB
}

func NewBboltStore(dbPath string) (ports.StorageService, error) {
	db, err := bbolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("no se pudo abrir la base de datos bbolt: %w", err)
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(historyBucket)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("no se pudo crear el bucket de historial: %w", err)
	}

	return &BboltStore{db: db}, nil
}

func (s *BboltStore) AddToHistory(entry domain.HistoryEntry) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(historyBucket)

		key, err := entry.PlayedAt.MarshalBinary()
		if err != nil {
			return err
		}

		value, err := json.Marshal(entry)
		if err != nil {
			return fmt.Errorf("error al serializar la entrada de historial: %w", err)
		}

		return b.Put(key, value)
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
				return fmt.Errorf("error al deserializar entrada de historial: %w", err)
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
