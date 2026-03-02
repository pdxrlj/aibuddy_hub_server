package flash

import (
	logger "aibuddy/pkg/log"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

var _ Flash = (*Redis)(nil)

// popScript is a Lua script that atomically gets and deletes a key.
var popScript = redis.NewScript(`
	local val = redis.call('GET', KEYS[1])
	if val then
		redis.call('DEL', KEYS[1])
	end
	return val
`)

// Redis is a Redis cache implementation.
type Redis struct {
	client *redis.Client
	prefix string
}

// RedisOption is a function that configures Redis.
type RedisOption func(*Redis)

// WithRedisClient sets an existing Redis client.
func WithRedisClient(client *redis.Client) RedisOption {
	return func(r *Redis) { r.client = client }
}

// WithRedisAddr sets Redis address.
func WithRedisAddr(addr string) RedisOption {
	return func(r *Redis) {
		r.client = redis.NewClient(&redis.Options{Addr: addr})
	}
}

// WithRedisConfig sets Redis configuration.
func WithRedisConfig(host string, port int, username, password string, db int) RedisOption {
	return func(r *Redis) {
		slog.Info(logger.Redis, "host", host, "port", port, "username", username, "password", password, "db", db)
		r.client = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", host, port),
			Password: password,
			DB:       db,
			Username: username,
		})
	}
}

// WithPrefix sets key prefix.
func WithPrefix(prefix string) RedisOption {
	return func(r *Redis) { r.prefix = prefix }
}

// NewRedis creates a new Redis cache.
func NewRedis(opts ...RedisOption) (*Redis, error) {
	r := &Redis{prefix: "flash"}
	for _, opt := range opts {
		opt(r)
	}
	if r.client == nil {
		return nil, fmt.Errorf("redis client is required")
	}
	if err := r.client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}
	return r, nil
}

func (r *Redis) key(key string) string {
	return r.prefix + ":" + key
}

// Set stores a value with optional TTL.
func (r *Redis) Set(key string, value any, ttl ...time.Duration) error {
	var d time.Duration
	if len(ttl) > 0 {
		d = ttl[0]
	}
	return r.client.Set(context.Background(), r.key(key), value, d).Err()
}

var updateOrInsertScript = redis.NewScript(`
	local currentTTL = redis.call('TTL', KEYS[1])
	if currentTTL == -2 then
		currentTTL = -1
	end
	if currentTTL == -1 then
		return redis.call('SET', KEYS[1], ARGV[1])
	else
		return redis.call('SET', KEYS[1], ARGV[1], 'EX', currentTTL)
	end
`)

// Upsert updates a value if it exists (preserving TTL), otherwise inserts it with optional TTL.
func (r *Redis) Upsert(key string, value any, ttl ...time.Duration) error {
	ctx := context.Background()
	fullKey := r.key(key)

	// key 不存在，直接插入
	if !r.Exists(key) {
		var d time.Duration
		if len(ttl) > 0 {
			d = ttl[0]
		}
		return r.client.Set(ctx, fullKey, value, d).Err()
	}

	// key 存在，更新值并保留原有 TTL
	return updateOrInsertScript.Run(ctx, r.client, []string{fullKey}, value).Err()
}

// Get retrieves a value by key.
func (r *Redis) Get(key string) (any, error) {
	return r.client.Get(context.Background(), r.key(key)).Result()
}

// Delete removes a key from the cache.
func (r *Redis) Delete(key string) error {
	return r.client.Del(context.Background(), r.key(key)).Err()
}

// Pop retrieves and removes a key atomically using Lua script.
func (r *Redis) Pop(key string) (any, error) {
	result, err := popScript.Run(context.Background(), r.client, []string{r.key(key)}).Result()
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, redis.Nil
	}
	return result, nil
}

// Exists checks if a key exists.
func (r *Redis) Exists(key string) bool {
	n, _ := r.client.Exists(context.Background(), r.key(key)).Result()
	return n > 0
}

// Incr atomically increments a key. If key doesn't exist, sets to 1 with optional TTL.
func (r *Redis) Incr(key string, ttl ...time.Duration) (int64, error) {
	ctx := context.Background()
	fullKey := r.key(key)

	// 原子递增
	result, err := r.client.Incr(ctx, fullKey).Result()
	if err != nil {
		return 0, err
	}

	// 首次创建时设置过期时间
	if result == 1 && len(ttl) > 0 && ttl[0] > 0 {
		if err := r.client.Expire(ctx, fullKey, ttl[0]).Err(); err != nil {
			return result, err
		}
	}

	return result, nil
}
