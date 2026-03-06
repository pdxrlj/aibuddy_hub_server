// Package file 文件服务
package file

import (
	"aibuddy/pkg/config"
	"aibuddy/pkg/storage"
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// tracer 获取 tracer
var tracer = func() trace.Tracer {
	return otel.Tracer(config.Instance.Tracer.ServiceName)
}

// Service 文件服务
type Service struct {
	FileStorage storage.ObjectStorage[io.ReadCloser]
}

// NewService 创建文件服务
func NewService() *Service {
	if config.Instance.Storage == nil || config.Instance.Storage.OSS == nil {
		panic("storage config is not set")
	}
	return &Service{
		FileStorage: storage.NewStorage(
			config.Instance.Storage.OSS.AccessKeyID,
			config.Instance.Storage.OSS.AccessKeySecret,
			config.Instance.Storage.OSS.Region,
			config.Instance.Storage.OSS.Endpoint,
			config.Instance.Storage.OSS.Bucket,
		),
	}
}

// UploadFile 上传文件
func (f *Service) UploadFile(ctx context.Context, deviceID string, file *multipart.FileHeader) (filename, presignedURL string, err error) {
	_, span := tracer().Start(ctx, "FileService.UploadFile")
	defer span.End()
	stream, err := file.Open()
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		return "", "", err
	}
	defer func() {
		_ = stream.Close()
	}()

	fileName := fmt.Sprintf("%s/%s", deviceID, file.Filename)
	if err = f.FileStorage.Storage(ctx, fileName, stream); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		return "", "", fmt.Errorf("上传文件失败: %w", err)
	}

	presignedURL, err = f.FileStorage.PresignURL(ctx, fileName, 15*time.Minute)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		return "", "", fmt.Errorf("生成预签名URL失败: %w", err)
	}

	slog.Info("[UploadFile]", "device_id", deviceID, "filename", fileName, "presigned_url", presignedURL)
	return fileName, presignedURL, nil
}

// FileProxy 文件代理
func (f *Service) FileProxy(ctx context.Context, deviceID, filename string) (io.ReadCloser, error) {
	_, span := tracer().Start(ctx, "FileService.FileProxy")
	defer span.End()

	file, err := f.FileStorage.Download(ctx, filename)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		return nil, err
	}
	return file, nil
}
