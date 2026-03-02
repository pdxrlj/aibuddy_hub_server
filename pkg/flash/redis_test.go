package flash

import (
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestRedis(t *testing.T) *Redis {
	addr := os.Getenv("REDIS_TEST_ADDR")
	if addr == "" {
		addr = "112.132.224.74:6379"
	}
	password := os.Getenv("REDIS_TEST_PASSWORD")

	var opts []RedisOption
	if password != "" {
		host := "112.132.224.74"
		port := 6379
		if addr != "" {
			parts := strings.Split(addr, ":")
			if len(parts) == 2 {
				host = parts[0]
				port, _ = strconv.Atoi(parts[1])
			}
		}
		opts = []RedisOption{
			WithRedisConfig(host, port, "", password, 0),
			WithPrefix("test"),
		}
	} else {
		opts = []RedisOption{
			WithRedisAddr(addr),
			WithPrefix("test"),
		}
	}

	r, err := NewRedis(opts...)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	return r
}

func TestRedis_UpdateOrInsert(t *testing.T) {
	r := getTestRedis(t)
	ctx := t.Context()

	tests := []struct {
		name     string
		setup    func(key string)
		key      string
		value    string
		ttl      time.Duration
		wantTTL  time.Duration
		checkTTL bool
	}{
		{
			name:     "insert new key without TTL",
			setup:    nil,
			key:      "new_key_1",
			value:    "value1",
			ttl:      0,
			wantTTL:  -1,
			checkTTL: true,
		},
		{
			name:     "insert new key with TTL",
			setup:    nil,
			key:      "new_key_2",
			value:    "value2",
			ttl:      10 * time.Second,
			wantTTL:  10 * time.Second,
			checkTTL: true,
		},
		{
			name: "update existing key without TTL",
			setup: func(key string) {
				_ = r.Set(key, "old_value", 0)
			},
			key:      "existing_key_1",
			value:    "new_value",
			ttl:      0,
			wantTTL:  -1,
			checkTTL: true,
		},
		{
			name: "update existing key with TTL - preserve TTL",
			setup: func(key string) {
				_ = r.Set(key, "old_value", 30*time.Second)
			},
			key:      "existing_key_2",
			value:    "new_value",
			ttl:      0,
			wantTTL:  30 * time.Second,
			checkTTL: true,
		},
		{
			name: "update existing key with TTL - ignore new TTL and preserve original",
			setup: func(key string) {
				_ = r.Set(key, "old_value", 30*time.Second)
			},
			key:      "existing_key_3",
			value:    "new_value",
			ttl:      60 * time.Second,
			wantTTL:  30 * time.Second, // 保留原 TTL，忽略传入的 TTL
			checkTTL: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := tt.key + "_" + time.Now().Format("20060102150405")

			if tt.setup != nil {
				tt.setup(key)
			}

			err := r.Upsert(key, tt.value, tt.ttl)
			require.NoError(t, err)

			got, err := r.Get(key)
			require.NoError(t, err)
			assert.Equal(t, tt.value, got)

			if tt.checkTTL && tt.wantTTL > 0 {
				ttl := r.client.TTL(ctx, r.key(key)).Val()
				assert.GreaterOrEqual(t, ttl.Seconds(), tt.wantTTL.Seconds()-1)
			}

			if tt.checkTTL && tt.wantTTL == -1 {
				ttl := r.client.TTL(ctx, r.key(key)).Val()
				assert.Equal(t, time.Duration(-1), ttl)
			}

			_ = r.Delete(key)
		})
	}
}

func TestRedis_UpdateOrInsert_PreserveTTL(t *testing.T) {
	r := getTestRedis(t)

	key := "preserve_ttl_test"
	originalTTL := 30 * time.Second

	err := r.Set(key, "original_value", originalTTL)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	err = r.Upsert(key, "updated_value")
	require.NoError(t, err)

	got, err := r.Get(key)
	require.NoError(t, err)
	assert.Equal(t, "updated_value", got)

	ctx := t.Context()
	remainingTTL := r.client.TTL(ctx, r.key(key)).Val()
	assert.Greater(t, remainingTTL.Seconds(), float64(28), "TTL should be preserved")
	assert.LessOrEqual(t, remainingTTL.Seconds(), float64(30), "TTL should not exceed original")

	_ = r.Delete(key)
}

func TestRedis_BasicOperations(t *testing.T) {
	r := getTestRedis(t)

	key := "basic_test_key"
	value := "test_value"

	assert.False(t, r.Exists(key))

	err := r.Set(key, value)
	require.NoError(t, err)

	assert.True(t, r.Exists(key))

	got, err := r.Get(key)
	require.NoError(t, err)
	assert.Equal(t, value, got)

	err = r.Delete(key)
	require.NoError(t, err)

	assert.False(t, r.Exists(key))
}
