// Package helpers 提供通用帮助函数
package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
	"unsafe"

	randv2 "math/rand/v2"
)

// Retry 带超时的重试函数
func Retry(times int, timeout time.Duration, fn func() error) error {
	var lastErr error

	for i := 0; i < times; i++ {
		done := make(chan error, 1)

		go func() {
			defer func() {
				if r := recover(); r != nil {
					done <- fmt.Errorf("panic: %v", r)
				}
			}()
			done <- fn()
		}()

		select {
		case err := <-done:
			if err == nil {
				return nil
			}
			lastErr = err
		case <-time.After(timeout):
			lastErr = fmt.Errorf("timeout after %v", timeout)
		}

		if i < times-1 {
			time.Sleep(time.Millisecond * 100 * time.Duration(i+1)) // 递增延迟
		}
	}

	return fmt.Errorf("retry %d times failed, last error: %v", times, lastErr)
}

// RetryWithCancelableContext 带可取消上下文的重试函数
func RetryWithCancelableContext(times int, timeout time.Duration, fn func(ctx context.Context) error) error {
	var lastErr error

	for i := 0; i < times; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)

		done := make(chan error, 1)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					done <- fmt.Errorf("panic: %v", r)
				}
			}()
			done <- fn(ctx)
		}()

		select {
		case err := <-done:
			cancel()
			if err == nil {
				return nil
			}
			lastErr = err
		case <-ctx.Done():
			cancel()
			lastErr = fmt.Errorf("timeout after %v", timeout)
		}

		if i < times-1 {
			time.Sleep(time.Millisecond * 100 * time.Duration(i+1))
		}
	}

	return fmt.Errorf("retry %d times failed, last error: %v", times, lastErr)
}

// SimpleRetry 简单重试函数
func SimpleRetry(times int, fn func() error) error {
	var lastErr error

	for i := 0; i < times; i++ {
		err := fn()
		if err == nil {
			return nil
		}
		lastErr = err

		// 如果不是最后一次重试，等待一段时间再重试
		if i < times-1 {
			time.Sleep(time.Millisecond * 100 * time.Duration(i+1))
		}
	}

	return fmt.Errorf("retry %d times failed, last error: %v", times, lastErr)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var src = rand.NewSource(time.Now().UnixNano())

const (
	letterIdxBits = 6
	letterIdxMask = 1<<letterIdxBits - 1
	letterIdxMax  = 63 / letterIdxBits
)

// GenerateUUID 生成指定长度的随机字符串
func GenerateUUID(n int) string {
	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

// PP 格式化打印任意值
func PP(v any) {
	data, err := json.MarshalIndent(v, "", "  		")
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(string(data))
}

// GenerateNumber 生成数字随机数
func GenerateNumber(n int) string {
	var sb strings.Builder
	for i := 0; i < n; i++ {
		sb.WriteString(fmt.Sprintf("%d", randv2.Int32N(10)))
	}
	return sb.String()
}
