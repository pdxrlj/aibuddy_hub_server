// Package file 文件服务
package file

import (
	"aibuddy/pkg/config"
	"aibuddy/pkg/helpers"
	"aibuddy/pkg/storage"
	"context"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"mime/multipart"
	"path"
	"time"

	ffmpeg "github.com/u2takey/ffmpeg-go"
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
func (f *Service) UploadFile(ctx context.Context, deviceID string, file *multipart.FileHeader, enableAudioTranscode bool, destAudioFormat string) (filename, presignedURL string, err error) {
	_, span := tracer().Start(ctx, "FileService.UploadFile")
	defer span.End()

	var stream io.ReadCloser
	stream, err = file.Open()
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		return "", "", err
	}
	defer func() {
		_ = stream.Close()
	}()

	fileName := fmt.Sprintf("%s/%s", deviceID, file.Filename)
	if enableAudioTranscode {
		ext := path.Ext(file.Filename)
		PcmParams := helpers.Cond(ext == ".pcm", defaultPCMParams(), nil)
		stream, err = f.AudioTranscode(ctx, stream, destAudioFormat, PcmParams)
		if err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("device_id", deviceID))
			return "", "", err
		}

		baseName := file.Filename[:len(file.Filename)-len(ext)]
		fileName = fmt.Sprintf("%s/%s.%s", deviceID, baseName, destAudioFormat)
	}

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

func defaultPCMParams() map[string]any {
	return map[string]any{
		"ar":     "16000",     // 采样率
		"ac":     "1",         // 声道数
		"f":      "s16le",     // 格式
		"acodec": "pcm_s16le", // 编码
	}
}

// AudioTranscode 音频转码
func (f *Service) AudioTranscode(
	ctx context.Context,
	src io.ReadCloser,
	destFormat string,
	pcmParams ...map[string]any,
) (io.ReadCloser, error) {
	_, span := tracer().Start(ctx, "FileService.AudioTranscode")
	defer span.End()

	pr, pw := io.Pipe()

	inputArgs := ffmpeg.KwArgs{}

	if len(pcmParams) > 0 {
		defaultPCMParams := defaultPCMParams()
		maps.Copy(defaultPCMParams, pcmParams[0])
		maps.Copy(inputArgs, defaultPCMParams)
	}

	cmd := ffmpeg.Input("pipe:0", inputArgs).Output("pipe:1", ffmpeg.KwArgs{
		"format":       destFormat,
		"acodec":       getAudioCodec(destFormat),
		"ar":           "16000",
		"ac":           "1",
		"loglevel":     "error",
		"hide_banner":  "",
		"map_metadata": "-1",
		"movflags":     "+faststart",
	}).WithInput(src).WithOutput(pw)

	go func() {
		defer func() {
			_ = src.Close()
			_ = pw.Close()
		}()

		if err := cmd.Run(); err != nil {
			slog.Error("AudioTranscode failed", "error", err)
			_ = pw.CloseWithError(fmt.Errorf("audio transcode failed: %w", err))
		}
	}()

	return pr, nil
}

// getAudioCodec 根据格式获取音频编码器
func getAudioCodec(format string) string {
	switch format {
	case "mp3":
		return "libmp3lame"
	case "aac", "m4a":
		return "aac"
	case "ogg":
		return "libvorbis"
	case "wav":
		return "pcm_s16le"
	case "flac":
		return "flac"
	case "opus":
		return "libopus"
	default:
		return "copy"
	}
}
