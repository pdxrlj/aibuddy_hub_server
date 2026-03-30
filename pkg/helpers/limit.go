package helpers

import (
	"errors"
	"time"

	"aibuddy/pkg/flash"
)

// ErrLimitExceeded 限流错误
var ErrLimitExceeded = errors.New("rate limit exceeded")

// Limiter 限流器
type Limiter struct {
	cache    flash.Flash   // 缓存实例
	maxCount int64         // 时间窗口内最大执行次数
	window   time.Duration // 时间窗口
}

// NewLimiter 创建限流器
// cache: 缓存实例
// maxCount: 时间窗口内最大执行次数
// window: 时间窗口
func NewLimiter(cache flash.Flash, maxCount int64, window time.Duration) *Limiter {
	return &Limiter{
		cache:    cache,
		maxCount: maxCount,
		window:   window,
	}
}

// Allow 检查是否允许执行
// key: 限流键，用于区分不同的限流对象
// 使用固定窗口算法：窗口从第一次请求开始计时，不随后续请求滑动
func (l *Limiter) Allow(key string) bool {
	// 先检查是否已存在，避免 TTL 被重置
	if !l.cache.Exists(key) {
		// 首次请求，设置计数为1并设置 TTL
		count, err := l.cache.Incr(key, l.window)
		if err != nil {
			return false
		}
		return count <= l.maxCount
	}

	// 已存在，只递增计数（不更新 TTL）
	count, err := l.cache.Incr(key)
	if err != nil {
		return false
	}
	return count <= l.maxCount
}

// Execute 执行函数（如果允许的话）
// key: 限流键
// fn: 要执行的函数
func (l *Limiter) Execute(key string, fn func() error) error {
	if !l.Allow(key) {
		return ErrLimitExceeded
	}
	return fn()
}

// Remaining 获取剩余可用次数
// key: 限流键
func (l *Limiter) Remaining(key string) int64 {
	val, err := l.cache.Get(key)
	if err != nil {
		return l.maxCount
	}
	count, ok := val.(int64)
	if !ok {
		return l.maxCount
	}
	remaining := l.maxCount - count
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Reset 重置限流计数
// key: 限流键
func (l *Limiter) Reset(key string) error {
	return l.cache.Delete(key)
}
