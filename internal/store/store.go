package store

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"
)

type entry struct {
	value     string
	expiresAt time.Time
}

type Store struct {
	mu     sync.Mutex
	saveMu sync.Mutex
	data   map[string]entry
}

func NewStore() *Store {
	return &Store{
		data: make(map[string]entry),
	}
}

func (s *Store) Set(key string, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = entry{
		value: value,
	}
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.getLiveEntryLocked(key)
	return item.value, exists
}

func (s *Store) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.getLiveEntryLocked(key)
	if !exists {
		return false
	}

	delete(s.data, key)
	return true
}

func (s *Store) DeleteMany(keys []string) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	var deleted int64

	for _, key := range keys {
		if _, exists := s.getLiveEntryLocked(key); exists {
			delete(s.data, key)
			deleted++
		}
	}

	return deleted
}

func (s *Store) Incr(key string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, exists := s.getLiveEntryLocked(key)
	if !exists {
		s.data[key] = entry{
			value: "1",
		}
		return 1, nil
	}

	parsedVal, err := strconv.ParseInt(val.value, 10, 64)
	if err != nil {
		return 0, errors.New("value is not a valid int64")
	}

	if parsedVal == math.MaxInt64 {
		return 0, errors.New("increment would overflow int64")
	}

	newValue := parsedVal + 1

	val.value = strconv.FormatInt(newValue, 10)
	s.data[key] = val

	return newValue, nil
}

func (s *Store) getLiveEntryLocked(key string) (entry, bool) {
	item, exists := s.data[key]
	if !exists {
		return entry{}, false
	}

	if !item.expiresAt.IsZero() && !time.Now().Before(item.expiresAt) {
		delete(s.data, key)
		return entry{}, false
	}

	return item, true
}

func (s *Store) Expire(key string, ttl time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, exists := s.getLiveEntryLocked(key)
	if !exists {
		return false
	}

	if ttl <= 0 {
		delete(s.data, key)
		return true
	}

	val.expiresAt = time.Now().Add(ttl)
	s.data[key] = val

	return true
}

func (s *Store) TTL(key string) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, exists := s.getLiveEntryLocked(key)
	if !exists {
		return -2
	}

	if val.expiresAt.IsZero() {
		return -1
	}

	remaining := time.Until(val.expiresAt)

	if remaining <= 0 {
		delete(s.data, key)
		return -2
	}

	return int64(remaining / time.Second)
}

func (s *Store) SetWithTTL(
	key string,
	value string,
	ttl time.Duration,
) error {
	if ttl <= 0 {
		return fmt.Errorf("ttl must be greater than zero")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = entry{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}

	return nil
}

// Cleans up the keys that are not readed (in the background)
func (s *Store) DeleteExpired() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	deleted := 0

	for key, val := range s.data {
		if !val.expiresAt.IsZero() && !now.Before(val.expiresAt) {
			delete(s.data, key)
			deleted++
		}
	}

	return deleted
}

func (s *Store) RunCleanup(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.DeleteExpired()

		case <-ctx.Done():
			return
		}
	}
}
