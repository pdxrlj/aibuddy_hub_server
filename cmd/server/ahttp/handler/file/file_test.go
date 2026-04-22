package filehandler

import (
	"aibuddy/pkg/config"
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileProxy(t *testing.T) {
	// 初始化配置
	cfg := config.Setup("f:/code/aibuddy_hub_server/config")
	t.Logf("config: %+v", cfg.Storage.OSS)

	// 创建 File handler
	fileHandler := NewFile()

	// 测试文件名
	filename := "30:ED:A0:E9:F3:07/7608691418.bin"

	t.Run("完整下载", func(t *testing.T) {
		ctx := context.Background()
		result, err := fileHandler.Service.FileProxy(ctx, "", filename, "", "")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		defer func() { _ = result.Body.Close() }()

		t.Logf("完整下载: ContentLength = %d, ContentType = %s", result.ContentLength, result.ContentType)

		// 读取全部内容
		data, err := io.ReadAll(result.Body)
		assert.NoError(t, err)
		t.Logf("完整下载: 文件大小 = %.2f MB", float64(len(data))/1024/1024)
	})

	t.Run("分片下载-前1024字节", func(t *testing.T) {
		ctx := context.Background()
		// Range: bytes=0-1023 获取前 1024 字节
		result, err := fileHandler.Service.FileProxy(ctx, "", filename, "bytes=0-1023", "")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		defer func() { _ = result.Body.Close() }()

		t.Logf("分片下载(0-1023): ContentLength = %d, ContentRange = %s", result.ContentLength, result.ContentRange)

		// 读取分片内容
		data, err := io.ReadAll(result.Body)
		assert.NoError(t, err)
		assert.LessOrEqual(t, len(data), 1024)
		t.Logf("分片下载(0-1023): 获取大小 = %.2f MB", float64(len(data))/1024/1024)
	})

	t.Run("分片下载-从1024开始到末尾", func(t *testing.T) {
		ctx := context.Background()
		// Range: bytes=1024- 从第 1024 字节开始到末尾
		result, err := fileHandler.Service.FileProxy(ctx, "", filename, "bytes=1024-", "")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		defer func() { _ = result.Body.Close() }()

		t.Logf("分片下载(1024-): ContentLength = %d, ContentRange = %s", result.ContentLength, result.ContentRange)

		data, err := io.ReadAll(result.Body)
		assert.NoError(t, err)
		t.Logf("分片下载(1024-): 获取大小 = %.2f MB", float64(len(data))/1024/1024)
	})

	t.Run("分片下载-最后500字节", func(t *testing.T) {
		ctx := context.Background()
		// Range: bytes=-500 获取最后 500 字节
		result, err := fileHandler.Service.FileProxy(ctx, "", filename, "bytes=-500", "")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		defer func() { _ = result.Body.Close() }()

		t.Logf("分片下载(-500): ContentLength = %d, ContentRange = %s", result.ContentLength, result.ContentRange)

		data, err := io.ReadAll(result.Body)
		assert.NoError(t, err)
		assert.LessOrEqual(t, len(data), 500)
		t.Logf("分片下载(-500): 获取大小 = %.2f MB", float64(len(data))/1024/1024)
	})

	t.Run("分片下载-中间范围", func(t *testing.T) {
		ctx := context.Background()
		// Range: bytes=1000-2000 获取中间部分
		result, err := fileHandler.Service.FileProxy(ctx, "", filename, "bytes=1000-2000", "")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		defer func() { _ = result.Body.Close() }()

		t.Logf("分片下载(1000-2000): ContentLength = %d, ContentRange = %s", result.ContentLength, result.ContentRange)

		data, err := io.ReadAll(result.Body)
		assert.NoError(t, err)
		t.Logf("分片下载(1000-2000): 获取大小 = %.2f MB", float64(len(data))/1024/1024)
	})
}
