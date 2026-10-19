package store

import (
	"errors"
	"math"
	"strconv"
	"sync"
)

type Store struct {
	mu   sync.Mutex
	data map[string]string
}

func NewStore() *Store {
	return &Store{
		data: make(map[string]string),
	}
}

func (s *Store) Set(key string, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = value
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, ok := s.data[key]
	return val, ok
}

func (s *Store) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.data[key]
	delete(s.data, key)
	return ok
}

func (s *Store) Incr(key string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, exists := s.data[key]
	if !exists {
		s.data[key] = "1"
		return 1, nil
	}

	parsedVal, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, errors.New("value is not a valid int64")
	}

	if parsedVal == math.MaxInt64 {
		return 0, errors.New("increment would overflow int64")
	}

	newValue := parsedVal + 1
	s.data[key] = strconv.FormatInt(newValue, 10)

	return newValue, nil
}
