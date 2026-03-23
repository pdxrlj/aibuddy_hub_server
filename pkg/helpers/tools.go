// Package helpers 提供通用帮助函数
package helpers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unsafe"

	randv2 "math/rand/v2"
)

// RetrySkippable 定义接口，实现此接口的错误将跳过重试
type RetrySkippable interface {
	SkipRetry() bool
}

// RetrySkipError 实现 RetrySkippable 接口，表示该错误不需要重试
type RetrySkipError struct {
	Message string
}

// Error 实现 error 接口
func (e RetrySkipError) Error() string {
	return e.Message
}

// SkipRetry 实现 RetrySkippable 接口，表示该错误不需要重试
func (e RetrySkipError) SkipRetry() bool {
	return true
}

// SkipRetry 跳过重试的错误
var SkipRetry = RetrySkipError{Message: "数据太少了,请过段时间再来吧"}

// IsSkipError 判断错误是否需要跳过重试
// 支持实现 RetrySkippable 接口的错误
func IsSkipError(err error) bool {
	if err == nil {
		return false
	}
	var s RetrySkippable
	if errors.As(err, &s) {
		return s.SkipRetry()
	}
	return false
}

// Retry 带超时的重试函数
func Retry[T any](times int, timeout time.Duration, fn func() (T, error)) (T, error) {
	var lastErr error
	var zero T

	for i := 0; i < times; i++ {
		done := make(chan struct {
			result T
			err    error
		}, 1)

		go func() {
			defer func() {
				if r := recover(); r != nil {
					done <- struct {
						result T
						err    error
					}{zero, fmt.Errorf("panic: %v", r)}
				}
			}()
			result, err := fn()
			done <- struct {
				result T
				err    error
			}{result, err}
			if i > 0 {
				slog.Info("[Helpers] Retry", "重试次数", i)
			}
		}()

		select {
		case res := <-done:
			if res.err == nil || IsSkipError(res.err) {
				return res.result, res.err
			}
			lastErr = res.err
		case <-time.After(timeout):
			lastErr = fmt.Errorf("timeout after %v", timeout)
		}

		if i < times-1 {
			time.Sleep(time.Millisecond * 100 * time.Duration(i+1)) // 递增延迟
		}
	}

	return zero, fmt.Errorf("retry %d times failed, last error: %v", times, lastErr)
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
		fmt.Fprintf(&sb, "%d", randv2.Int32N(10))
	}
	return sb.String()
}

// ValidateMobile 验证手机号码格式
func ValidateMobile(phone string) error {
	mobileRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
	if ok := mobileRegex.MatchString(phone); !ok {
		return errors.New("手机号格式不合法")
	}

	return nil
}

// Cond 三元运算辅助函数
func Cond[T any](condition bool, trueVal, falseVal T) T {
	if condition {
		return trueVal
	}
	return falseVal
}

// CompareVersion 比较两个版本号
func CompareVersion(v1, v2 string) int {
	v1Parts := strings.Split(v1, ".")
	v2Parts := strings.Split(v2, ".")

	maxLen := len(v1Parts)
	if len(v2Parts) > maxLen {
		maxLen = len(v2Parts)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(v1Parts) {
			n1, _ = strconv.Atoi(v1Parts[i])
		}
		if i < len(v2Parts) {
			n2, _ = strconv.Atoi(v2Parts[i])
		}

		if n1 < n2 {
			return -1
		} else if n1 > n2 {
			return 1
		}
	}
	return 0
}
