// Package filehandler 文件处理
package filehandler

import (
	"aibuddy/internal/services/file"
	"aibuddy/pkg/ahttp"
	"aibuddy/pkg/config"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = func() trace.Tracer {
	return otel.Tracer(config.Instance.Tracer.ServiceName)
}

// File 文件处理
type File struct {
	Service *file.Service
}

// NewFile 创建文件处理服务
func NewFile() *File {
	return &File{
		Service: file.NewService(),
	}
}

// UploadFile 上传文件（表单方式）
func (f *File) UploadFile(state *ahttp.State, req *UploadFileRequest) error {
	ctx, span := tracer().Start(state.Context(), "File.UploadFile")
	defer span.End()

	if req.File.Size >= 3<<20 {
		span.RecordError(errors.New("文件大小不能超过3MB"))
		span.SetAttributes(attribute.String("device_id", req.DeviceID))
		return state.Resposne().Error(errors.New("文件大小不能超过3MB"))
	}

	filename, presignedURL, err := f.Service.UploadFile(ctx, req.DeviceID, req.File, req.EnableAudioTranscode, req.DestAudioFormat)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", req.DeviceID))
		return state.Resposne().Error(err)
	}

	return state.Resposne().Success(&UploadFileResponse{
		Filename:     filename,
		PresignedURL: presignedURL,
	})
}

// UploadStream 流式上传文件
func (f *File) UploadStream(state *ahttp.State, req *UploadStreamRequest) error {
	ctx, span := tracer().Start(state.Context(), "File.UploadStream")
	defer span.End()

	body := state.Ctx.Request().Body
	defer body.Close()

	filename, presignedURL, err := f.Service.UploadStream(ctx, req.DeviceID, req.Ext, body, req.EnableAudioTranscode, req.DestAudioFormat)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", req.DeviceID))
		return state.Resposne().Error(err)
	}

	return state.Resposne().Success(&UploadFileResponse{
		Filename:     filename,
		PresignedURL: presignedURL,
	})
}

// FileProxy 文件代理
func (f *File) FileProxy(state *ahttp.State, req *FileProxyRequest) error {
	ctx, span := tracer().Start(state.Context(), "File.FileProxy")
	defer span.End()

	file, err := f.Service.FileProxy(ctx, req.DeviceID, req.Filename)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", req.DeviceID))
		return state.Resposne().Error(err)
	}
	defer func() {
		_ = file.Close()
	}()

	return state.Resposne().File(file, req.Filename)
}
