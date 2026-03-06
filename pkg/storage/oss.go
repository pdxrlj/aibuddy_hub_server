// Package storage 存储相关
package storage

import (
	"context"
	"io"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

// ObjectStorage 定义存储接口
var _ ObjectStorage[io.ReadCloser] = (*OSS)(nil)

// OSS 阿里云 OSS 存储
type OSS struct {
	Client *oss.Client
	bucket string
}

// NewOss 创建 OSS 存储实例
func NewOss(accessKeyID, accessKeySecret, region, endpoint, bucket string) *OSS {
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKeyID, accessKeySecret, ""),
		).
		WithRegion(region).
		WithEndpoint(endpoint)

	client := oss.NewClient(cfg)

	return &OSS{
		Client: client,
		bucket: bucket,
	}
}

// Storage 存储文件
func (o *OSS) Storage(ctx context.Context, key string, r io.Reader) error {
	_, err := o.Client.PutObject(ctx, &oss.PutObjectRequest{
		Bucket: oss.Ptr(o.bucket),
		Key:    oss.Ptr(key),
		Body:   r,
	})

	if err != nil {
		return err
	}

	return nil
}

// PresignURL 获取预签名 URL
func (o *OSS) PresignURL(ctx context.Context, key string, expires time.Duration, process ...string) (string, error) {
	result, err := o.Client.Presign(ctx, &oss.GetObjectRequest{
		Bucket: oss.Ptr(o.bucket),
		Key:    oss.Ptr(key),
		Process: func() *string {
			if len(process) > 0 {
				return oss.Ptr(process[0])
			}
			return nil
		}(),
	},
		oss.PresignExpires(expires),
	)

	if err != nil {
		return "", err
	}

	return result.URL, nil
}

// Download 下载文件（流式）
func (o *OSS) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	result, err := o.Client.GetObject(ctx, &oss.GetObjectRequest{
		Bucket: oss.Ptr(o.bucket),
		Key:    oss.Ptr(key),
	})

	if err != nil {
		return nil, err
	}

	return result.Body, nil
}

// Delete 删除文件
func (o *OSS) Delete(ctx context.Context, key string) error {
	_, err := o.Client.DeleteObject(ctx, &oss.DeleteObjectRequest{
		Bucket: oss.Ptr(o.bucket),
		Key:    oss.Ptr(key),
	})

	if err != nil {
		return err
	}

	return nil
}
