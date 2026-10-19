package store

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type snapshotEntry struct {
	Value     string    `json:"value"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (s *Store) snapshot() map[string]snapshotEntry {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	result := make(map[string]snapshotEntry)

	for key, val := range s.data {
		if !val.expiresAt.IsZero() && !now.Before(val.expiresAt) {
			continue
		}

		result[key] = snapshotEntry{
			Value:     val.value,
			ExpiresAt: val.expiresAt,
		}

	}

	return result
}

func (s *Store) SaveSnapshot(path string) error {
	s.saveMu.Lock()
	defer s.saveMu.Unlock()

	data := s.snapshot()

	// Create the temporary file in the destination directory.
	file, err := os.CreateTemp(filepath.Dir(path), ".snapshot-*")
	if err != nil {
		return fmt.Errorf("create snapshot file: %w", err)
	}

	tempPath := file.Name()
	defer os.Remove(tempPath)
	defer file.Close()

	if err := json.NewEncoder(file).Encode(data); err != nil {
		return fmt.Errorf("encode snapshot: %w", err)
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync snapshot: %w", err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("close snapshot: %w", err)
	}

	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("replace snapshot: %w", err)
	}

	return nil

}

func (s *Store) LoadSnapshot(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open snapshot: %w", err)
	}
	defer file.Close()

	var savedSnap map[string]snapshotEntry
	decoder := json.NewDecoder(file)

	if err := decoder.Decode(&savedSnap); err != nil {
		return fmt.Errorf("decode snapshot: %w", err)
	}

	// Reject extra JSON values or invalid trailing content.
	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("snapshot contains extra JSON data")
		}
		return fmt.Errorf("invalid trailing snapshot data: %w", err)
	}

	if savedSnap == nil {
		return fmt.Errorf("snapshot must be a JSON object")
	}

	restored := make(map[string]entry)
	now := time.Now()

	for key, item := range savedSnap {
		if !item.ExpiresAt.IsZero() && !now.Before(item.ExpiresAt) {
			continue
		}

		restored[key] = entry{
			value:     item.Value,
			expiresAt: item.ExpiresAt,
		}
	}

	s.mu.Lock()
	s.data = restored
	s.mu.Unlock()

	return nil
}
