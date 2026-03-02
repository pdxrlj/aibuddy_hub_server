package flash

import (
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto/v2"
)

var _ Flash = (*Memory)(nil)

// Memory is an in-memory cache implementation using ristretto.
type Memory struct {
	store *ristretto.Cache[string, any]
}

// NewMemory creates a new in-memory cache.
func NewMemory() (*Memory, error) {
	store, err := ristretto.NewCache(&ristretto.Config[string, any]{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create memory cache: %w", err)
	}
	return &Memory{store: store}, nil
}

// Set stores a value with optional TTL.
func (m *Memory) Set(key string, value any, ttl ...time.Duration) error {
	var d time.Duration
	if len(ttl) > 0 {
		d = ttl[0]
	}
	if !m.store.SetWithTTL(key, value, 1, d) {
		return fmt.Errorf("%w: %s", ErrSetFailed, key)
	}
	return nil
}

// Upsert updates a value if it exists, otherwise inserts it.
func (m *Memory) Upsert(key string, value any, ttl ...time.Duration) error {
	if len(ttl) > 0 && ttl[0] > 0 {
		return m.Set(key, value, ttl[0])
	}

	if m.Exists(key) {
		existingTTL, hasTTL := m.store.GetTTL(key)
		if hasTTL {
			return m.Set(key, value, existingTTL)
		}
		return m.Set(key, value)
	}
	return m.Set(key, value, ttl...)
}

// Get retrieves a value by key.
func (m *Memory) Get(key string) (any, error) {
	val, ok := m.store.Get(key)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}
	return val, nil
}

// Delete removes a key from the cache.
func (m *Memory) Delete(key string) error {
	m.store.Del(key)
	return nil
}

// Pop retrieves and removes a key atomically.
func (m *Memory) Pop(key string) (any, error) {
	val, ok := m.store.Get(key)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}
	m.store.Del(key)
	return val, nil
}

// Exists checks if a key exists.
func (m *Memory) Exists(key string) bool {
	_, ok := m.store.Get(key)
	return ok
}
