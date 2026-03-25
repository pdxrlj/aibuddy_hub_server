package storage

import (
	"context"
	"io"
	"time"
)

// ObjectStorage 定义对象存储接口
type ObjectStorage[T any] interface {
	PresignURL(ctx context.Context, key string, expires time.Duration, process ...string) (string, error)
	Download(ctx context.Context, key string, bytesRange string, process ...string) (T, error)
	Delete(ctx context.Context, key string) error
	Storage(ctx context.Context, key string, r io.Reader) error
}

// NewStorage 创建存储实例
func NewStorage(accessKeyID, accessKeySecret, region, endpoint, bucket string) ObjectStorage[io.ReadCloser] {
	return NewOss(accessKeyID, accessKeySecret, region, endpoint, bucket)
}
