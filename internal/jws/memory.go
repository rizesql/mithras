package jws

import (
	"context"
	"sync"
)

// MemoryStore is an in-memory implementation of the Store interface.
type MemoryStore struct {
	key Key
	mu  sync.RWMutex
}

// NewMemoryStore creates a new MemoryStore with a single generated key.
func NewMemoryStore(keygen KeyGenerator) (*MemoryStore, error) {
	key, err := keygen()
	if err != nil {
		return nil, err
	}

	return &MemoryStore{key: key}, nil
}

func (s *MemoryStore) SigningKey(_ context.Context) (Key, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.key, nil
}

func (s *MemoryStore) PublicKeys(_ context.Context) ([]Key, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return []Key{s.key}, nil
}
