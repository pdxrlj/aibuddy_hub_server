package storage

import (
	"context"
	"io"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
)

// ObjectStorage 定义对象存储接口
type ObjectStorage[T any] interface {
	PresignURL(ctx context.Context, key string, expires time.Duration, process ...string) (string, error)
	Download(ctx context.Context, key string) (T, error)
	Delete(ctx context.Context, key string) error
	Storage(ctx context.Context, key string, r io.Reader) error
}

// NewStorage 创建存储实例
func NewStorage(accessKeyID, accessKeySecret, region, endpoint, bucket string) ObjectStorage[*oss.GetObjectResult] {
	return NewOss(accessKeyID, accessKeySecret, region, endpoint, bucket)
}
