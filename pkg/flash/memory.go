package flash

import (
	"fmt"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto/v2"
)

var _ Flash = (*Memory)(nil)

// Memory is an in-memory cache implementation using ristretto.
type Memory struct {
	store *ristretto.Cache[string, any]
	mu    sync.Mutex // 用于 Incr 的原子操作
}

// NewMemory creates a new in-memory cache.
func NewMemory() (Flash, error) {
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

// Upsert updates a value if it exists (preserving TTL), otherwise inserts it with optional TTL.
func (m *Memory) Upsert(key string, value any, ttl ...time.Duration) error {
	// key 存在，更新值并保留原有 TTL
	if m.Exists(key) {
		existingTTL, hasTTL := m.store.GetTTL(key)
		if hasTTL {
			return m.Set(key, value, existingTTL)
		}
		return m.Set(key, value)
	}

	// key 不存在，插入新值
	var d time.Duration
	if len(ttl) > 0 {
		d = ttl[0]
	}
	return m.Set(key, value, d)
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

// Incr atomically increments a key. If key doesn't exist, sets to 1 with optional TTL.
func (m *Memory) Incr(key string, ttl ...time.Duration) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var count int64
	if val, ok := m.store.Get(key); ok {
		if n, ok := val.(int64); ok {
			count = n
		}
	}

	count++

	var d time.Duration
	if len(ttl) > 0 {
		d = ttl[0]
	}

	if !m.store.SetWithTTL(key, count, 1, d) {
		return 0, fmt.Errorf("%w: %s", ErrSetFailed, key)
	}

	// 等待异步写入完成
	m.store.Wait()

	return count, nil
}

// TTL returns the remaining TTL of a key.
// Returns (0, false) if key doesn't exist or has no TTL.
func (m *Memory) TTL(key string) (time.Duration, bool) {
	ttl, hasTTL := m.store.GetTTL(key)
	if !hasTTL || ttl <= 0 {
		return 0, false
	}
	return ttl, true
}
