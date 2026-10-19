package store

import (
	"errors"
	"math"
	"strconv"
)

type Store struct {
	data map[string]string
}

func NewStore() *Store {
	return &Store{
		data: make(map[string]string),
	}
}

func (s *Store) Set(key string, value string) {
	s.data[key] = value
}

func (s *Store) Get(key string) (string, bool) {
	val, ok := s.data[key]
	return val, ok
}

func (s *Store) Delete(key string) bool {
	_, ok := s.data[key]
	delete(s.data, key)
	return ok
}

func (s *Store) Incr(key string) (int64, error) {
	val, exists := s.Get(key)
	if !exists {
		s.Set(key, "1")
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
	s.Set(key, strconv.FormatInt(newValue, 10))

	return newValue, nil
}
