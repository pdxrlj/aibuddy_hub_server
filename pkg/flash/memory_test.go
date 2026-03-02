package flash

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemory_UpdateOrInsert(t *testing.T) {
	m, err := NewMemory()
	require.NoError(t, err)

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
			wantTTL:  0,
			checkTTL: false,
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
				_ = m.Set(key, "old_value", 0)
			},
			key:      "existing_key_1",
			value:    "new_value",
			ttl:      0,
			wantTTL:  0,
			checkTTL: false,
		},
		{
			name: "update existing key with TTL - preserve TTL",
			setup: func(key string) {
				_ = m.Set(key, "old_value", 30*time.Second)
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
				_ = m.Set(key, "old_value", 30*time.Second)
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
			key := tt.key + "_" + time.Now().Format("20060102150405.000000000")

			if tt.setup != nil {
				tt.setup(key)
				time.Sleep(10 * time.Millisecond)
			}

			err := m.Upsert(key, tt.value, tt.ttl)
			require.NoError(t, err)

			time.Sleep(10 * time.Millisecond)

			got, ok := m.store.Get(key)
			require.True(t, ok, "key should exist")
			assert.Equal(t, tt.value, got)

			if tt.checkTTL {
				remainingTTL, hasTTL := m.store.GetTTL(key)
				if tt.wantTTL > 0 {
					assert.True(t, hasTTL, "should have TTL")
					assert.GreaterOrEqual(t, remainingTTL.Milliseconds(), (tt.wantTTL - time.Second).Milliseconds(), "TTL should be preserved")
				}
			}

			m.Delete(key)
		})
	}
}

func TestMemory_UpdateOrInsert_PreserveTTL(t *testing.T) {
	m, err := NewMemory()
	require.NoError(t, err)

	key := "preserve_ttl_test"
	originalTTL := 30 * time.Second

	err = m.Set(key, "original_value", originalTTL)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	err = m.Upsert(key, "updated_value")
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	got, ok := m.store.Get(key)
	require.True(t, ok)
	assert.Equal(t, "updated_value", got)

	remainingTTL, hasTTL := m.store.GetTTL(key)
	assert.True(t, hasTTL, "TTL should be preserved")
	assert.Greater(t, remainingTTL.Milliseconds(), int64(28*1000), "TTL should be around 30 seconds")
	assert.LessOrEqual(t, remainingTTL.Milliseconds(), int64(30*1000), "TTL should not exceed original")

	m.Delete(key)
}

func TestMemory_BasicOperations(t *testing.T) {
	m, err := NewMemory()
	require.NoError(t, err)

	key := "basic_test_key"
	value := "test_value"

	assert.False(t, m.Exists(key))

	err = m.Set(key, value)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	assert.True(t, m.Exists(key))

	got, err := m.Get(key)
	require.NoError(t, err)
	assert.Equal(t, value, got)

	err = m.Delete(key)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	assert.False(t, m.Exists(key))
}

func TestMemory_Incr(t *testing.T) {
	m, err := NewMemory()
	require.NoError(t, err)

	t.Run("increment new key", func(t *testing.T) {
		key := "incr_new_" + time.Now().Format("20060102150405.000000000")

		count, err := m.Incr(key, 10*time.Second)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)

		time.Sleep(10 * time.Millisecond)

		// 验证 TTL 已设置
		ttl, hasTTL := m.store.GetTTL(key)
		assert.True(t, hasTTL)
		assert.Greater(t, ttl.Milliseconds(), int64(8000))

		m.Delete(key)
	})

	t.Run("increment existing key", func(t *testing.T) {
		key := "incr_existing_" + time.Now().Format("20060102150405.000000000")

		// 第一次递增
		count, err := m.Incr(key, 10*time.Second)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)

		// 第二次递增
		count, err = m.Incr(key)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)

		// 第三次递增
		count, err = m.Incr(key)
		require.NoError(t, err)
		assert.Equal(t, int64(3), count)

		time.Sleep(10 * time.Millisecond)

		// 验证值
		val, ok := m.store.Get(key)
		require.True(t, ok)
		assert.Equal(t, int64(3), val)

		m.Delete(key)
	})

	t.Run("increment without TTL", func(t *testing.T) {
		key := "incr_no_ttl_" + time.Now().Format("20060102150405.000000000")

		count, err := m.Incr(key)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)

		time.Sleep(10 * time.Millisecond)

		// 验证值正确
		val, ok := m.store.Get(key)
		require.True(t, ok)
		assert.Equal(t, int64(1), val)

		m.Delete(key)
	})

	t.Run("concurrent increment", func(t *testing.T) {
		key := "incr_concurrent_" + time.Now().Format("20060102150405.000000000")
		iterations := 100

		// 并发递增
		done := make(chan bool)
		for i := 0; i < iterations; i++ {
			go func() {
				_, _ = m.Incr(key)
				done <- true
			}()
		}

		// 等待所有 goroutine 完成
		for i := 0; i < iterations; i++ {
			<-done
		}

		time.Sleep(50 * time.Millisecond)

		// 验证最终值
		val, ok := m.store.Get(key)
		require.True(t, ok)
		assert.Equal(t, int64(iterations), val)

		m.Delete(key)
	})
}
