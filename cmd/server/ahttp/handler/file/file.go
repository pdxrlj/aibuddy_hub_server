// Package filehandler 文件处理
package filehandler

import (
	"aibuddy/internal/services/file"
	"aibuddy/pkg/ahttp"
	"aibuddy/pkg/config"
	"errors"
	"log/slog"
	"net/http"

	"github.com/spf13/cast"
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

	if req.File.Size >= 20<<20 {
		span.RecordError(errors.New("文件大小不能超过20MB"))
		span.SetAttributes(attribute.String("device_id", req.DeviceID))
		return state.Response().Error(errors.New("文件大小不能超过20MB"))
	}

	filename, presignedURL, err := f.Service.UploadFile(ctx, req.DeviceID, req.File, req.EnableAudioTranscode, req.DestAudioFormat)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", req.DeviceID))
		return state.Response().Error(err)
	}

	return state.Response().Success(&UploadFileResponse{
		Filename:     filename,
		PresignedURL: presignedURL,
	})
}

// UploadFileNoDeviceID 上传文件没有DeviceID
func (f *File) UploadFileNoDeviceID(state *ahttp.State, req *UploadFileNoDeviceIDRequest) error {
	ctx, span := tracer().Start(state.Context(), "File.UploadFileNoDeviceID")
	defer span.End()

	filename, presignedURL, err := f.Service.UploadFile(ctx, "", req.File, req.EnableAudioTranscode, req.DestAudioFormat)
	if err != nil {
		span.RecordError(err)
		return state.Response().Error(err)
	}

	slog.Info("[UploadFileNoDeviceID] 文件上传成功")

	return state.Response().Success(&UploadFileResponse{
		Filename:     filename,
		PresignedURL: presignedURL,
	})
}

// UploadStream 流式上传文件
func (f *File) UploadStream(state *ahttp.State, req *UploadStreamRequest) error {
	ctx, span := tracer().Start(state.Context(), "File.UploadStream")
	defer span.End()

	body := state.Ctx.Request().Body
	defer func() {
		_ = body.Close()
	}()

	filename, presignedURL, err := f.Service.UploadStream(ctx, req.DeviceID, req.Ext, body, req.EnableAudioTranscode, req.DestAudioFormat)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", req.DeviceID))
		return state.Response().Error(err)
	}

	slog.Info("[UploadStream] 流式文件上传成功")

	return state.Response().Success(&UploadFileResponse{
		Filename:     filename,
		PresignedURL: presignedURL,
	})
}

// FileProxy 文件代理
func (f *File) FileProxy(state *ahttp.State, req *FileProxyRequest) error {
	ctx, span := tracer().Start(state.Context(), "File.FileProxy")
	defer span.End()

	// 获取 Range header
	bytesRange := state.Ctx.Request().Header.Get("Range")

	result, err := f.Service.FileProxy(ctx, req.DeviceID, req.Filename, bytesRange, req.Process)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", req.DeviceID))
		return state.Response().Error(err)
	}
	defer func() {
		_ = result.Body.Close()
	}()

	resp := state.Response().SetHeaders(func() map[string]string {
		headers := map[string]string{}
		if result.ContentLength > 0 {
			headers["Content-Length"] = cast.ToString(result.ContentLength)
		}
		if result.ContentRange != "" {
			headers["Content-Range"] = result.ContentRange
		}
		return headers
	}())
	if result.ContentRange != "" {
		resp.SetStatus(http.StatusPartialContent)
	}
	return resp.File(result.Body, req.Filename)
}
