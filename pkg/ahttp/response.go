package ahttp

import (
	"io"
	"net/http"
	"path/filepath"

	"github.com/labstack/echo/v4"
)

// Response HTTP 响应结构
type Response struct {
	Data    any          `json:"data"`
	Ctx     echo.Context `json:"-"`
	Message string       `json:"message"`
	Status  int          `json:"status"`
}

// NewResponse 创建响应
func NewResponse(ctx echo.Context) *Response {
	return &Response{
		Ctx:     ctx,
		Status:  http.StatusOK,
		Message: http.StatusText(http.StatusOK),
		Data:    nil,
	}
}

// SetStatus 设置状态码
func (r *Response) SetStatus(status int) *Response {
	r.Status = status
	return r
}

// SetMessage 设置消息
func (r *Response) SetMessage(message string) *Response {
	r.Message = message
	return r
}

// SetData 设置数据
func (r *Response) SetData(data any) *Response {
	r.Data = data
	return r
}

// Success 返回成功响应
func (r *Response) Success(data ...any) error {
	if r.Ctx.Response().Committed {
		return nil
	}

	if r.Status != 0 {
		r.Status = http.StatusOK
	}
	if r.Message == "" {
		r.Message = http.StatusText(http.StatusOK)
	}
	if len(data) > 0 {
		r.Data = data[0]
	}
	return r.Ctx.JSON(http.StatusOK, r)
}

// NoContent 返回无内容响应
func (r *Response) NoContent() error {
	if r.Ctx.Response().Committed {
		return nil
	}
	if r.Status != 0 {
		r.Status = http.StatusOK
	}
	if r.Message == "" {
		r.Message = http.StatusText(http.StatusOK)
	}
	r.Message = http.StatusText(http.StatusOK)
	r.Data = nil
	return r.Ctx.JSON(http.StatusOK, r)
}

// Error 返回错误响应
func (r *Response) Error(err error) error {
	if r.Ctx.Response().Committed {
		return nil
	}
	if r.Status == http.StatusOK {
		r.Status = http.StatusInternalServerError
	}
	r.Message = err.Error()
	return r.Ctx.JSON(http.StatusOK, r)
}

// File 返回文件响应（流式）
func (r *Response) File(reader io.Reader, filename string) error {
	if r.Ctx.Response().Committed {
		return nil
	}
	// 根据文件扩展名获取 MIME 类型
	contentType := "application/octet-stream"
	if ext := filepath.Ext(filename); ext != "" {
		if mime := mimeTypes[ext]; mime != "" {
			contentType = mime
		}
	}
	r.Ctx.Response().Header().Set(echo.HeaderContentType, contentType)
	r.Ctx.Response().Header().Set(echo.HeaderContentDisposition, "inline; filename="+filepath.Base(filename))
	_, err := io.Copy(r.Ctx.Response(), reader)
	return err
}

// mimeTypes 文件扩展名到 MIME 类型的映射
var mimeTypes = map[string]string{
	".html": "text/html",
	".htm":  "text/html",
	".css":  "text/css",
	".js":   "application/javascript",
	".json": "application/json",
	".xml":  "application/xml",
	".txt":  "text/plain",
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
	".svg":  "image/svg+xml",
	".webp": "image/webp",
	".mp3":  "audio/mpeg",
	".wav":  "audio/wav",
	".mp4":  "video/mp4",
	".webm": "video/webm",
	".pdf":  "application/pdf",
	".zip":  "application/zip",
}
