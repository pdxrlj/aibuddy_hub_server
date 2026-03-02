// Package flash provides a unified cache interface for memory and Redis.
package flash

import (
	"errors"
	"fmt"
	"time"
)

// ErrKeyNotFound is returned when a key does not exist.
var ErrKeyNotFound = errors.New("key not found")

// ErrSetFailed is returned when a set operation fails.
var ErrSetFailed = errors.New("set flash failed")

// Flash defines the cache interface.
type Flash interface {
	Set(key string, value any, ttl ...time.Duration) error
	Upsert(key string, value any, ttl ...time.Duration) error
	Get(key string) (any, error)
	Delete(key string) error
	Pop(key string) (any, error) // Get and delete
	Exists(key string) bool
	TTL(key string) (time.Duration, bool)                 // 获取 TTL，返回剩余时间和是否存在 TTL
	Incr(key string, ttl ...time.Duration) (int64, error) // 原子递增，首次设置 TTL
}

// New creates a new Flash instance.
// driver: memory or redis
// opts: Redis options
func New(driver string, opts ...RedisOption) (Flash, error) {
	switch driver {
	case "memory":
		return NewMemory()
	case "redis":
		return NewRedis(opts...)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}
}
