package ahttp

import (
	"net/http"

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
