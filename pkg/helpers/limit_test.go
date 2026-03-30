package helpers

import (
	"errors"
	"testing"
	"time"

	"aibuddy/pkg/flash"
)

func TestLimiter_Execute(t *testing.T) {
	cache, err := flash.NewMemory()
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}

	// 创建限流器：100ms 内最多执行 3 次
	limiter := NewLimiter(cache, 3, 100*time.Millisecond)

	key := "test:execute"

	tests := []struct {
		name      string
		setup     func()
		wantErr   error
		callCount int // 连续调用次数
	}{
		{
			name:      "第一次调用应该成功",
			setup:     func() {},
			callCount: 1,
			wantErr:   nil,
		},
		{
			name: "在限制内的调用应该成功",
			setup: func() {
				// 重置计数
				limiter.Reset(key)
			},
			callCount: 3,
			wantErr:   nil,
		},
		{
			name: "超出限制应该返回 ErrLimitExceeded",
			setup: func() {
				limiter.Reset(key)
			},
			callCount: 4,
			wantErr:   ErrLimitExceeded,
		},
		{
			name: "时间窗口过期后应该重置计数",
			setup: func() {
				limiter.Reset(key)
				// 执行 3 次用完额度
				for i := 0; i < 3; i++ {
					limiter.Execute(key, func() error { return nil })
				}
				// 等待时间窗口过期
				time.Sleep(150 * time.Millisecond)
			},
			callCount: 1,
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			var lastErr error
			for i := 0; i < tt.callCount; i++ {
				lastErr = limiter.Execute(key, func() error {
					return nil
				})
			}

			if !errors.Is(lastErr, tt.wantErr) {
				t.Errorf("Execute() error = %v, want %v", lastErr, tt.wantErr)
			}
		})
	}
}

func TestLimiter_Execute_WithFnError(t *testing.T) {
	cache, err := flash.NewMemory()
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}

	limiter := NewLimiter(cache, 3, 100*time.Millisecond)
	key := "test:fn_error"

	fnErr := errors.New("function error")
	err = limiter.Execute(key, func() error {
		return fnErr
	})

	if !errors.Is(err, fnErr) {
		t.Errorf("Execute() should return function error, got %v", err)
	}
}

func TestLimiter_Allow(t *testing.T) {
	cache, err := flash.NewMemory()
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}

	limiter := NewLimiter(cache, 2, 100*time.Millisecond)
	key := "test:allow"

	// 第一次应该允许
	if !limiter.Allow(key) {
		t.Error("first call should be allowed")
	}

	// 第二次应该允许
	if !limiter.Allow(key) {
		t.Error("second call should be allowed")
	}

	// 第三次不应该允许
	if limiter.Allow(key) {
		t.Error("third call should not be allowed")
	}
}

func TestLimiter_Remaining(t *testing.T) {
	cache, err := flash.NewMemory()
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}

	limiter := NewLimiter(cache, 5, 100*time.Millisecond)
	key := "test:remaining"

	// 初始应该有 5 次
	if r := limiter.Remaining(key); r != 5 {
		t.Errorf("remaining should be 5, got %d", r)
	}

	// 执行一次后应该有 4 次
	limiter.Allow(key)
	if r := limiter.Remaining(key); r != 4 {
		t.Errorf("remaining should be 4, got %d", r)
	}

	// 执行两次后应该有 3 次
	limiter.Allow(key)
	if r := limiter.Remaining(key); r != 3 {
		t.Errorf("remaining should be 3, got %d", r)
	}
}

func TestLimiter_FixedWindow(t *testing.T) {
	cache, err := flash.NewMemory()
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}

	// 200ms 窗口，最多 3 次
	limiter := NewLimiter(cache, 3, 200*time.Millisecond)
	key := "test:fixed_window"

	// 第一次请求，设置 TTL
	limiter.Allow(key)

	// 获取初始 TTL
	ttl1, hasTTL1 := cache.TTL(key)
	if !hasTTL1 {
		t.Fatal("TTL should be set after first request")
	}

	// 等待 50ms
	time.Sleep(50 * time.Millisecond)

	// 第二次请求
	limiter.Allow(key)

	// TTL 应该减少约 50ms，而不是被重置
	ttl2, hasTTL2 := cache.TTL(key)
	if !hasTTL2 {
		t.Fatal("TTL should still exist")
	}

	// TTL 应该减少了大约 50ms，允许 10ms 误差
	expectedTTL := ttl1 - 50*time.Millisecond
	if ttl2 > expectedTTL+10*time.Millisecond || ttl2 < expectedTTL-10*time.Millisecond {
		t.Errorf("TTL should decrease, ttl1=%v, ttl2=%v, expected~%v", ttl1, ttl2, expectedTTL)
	}
}
