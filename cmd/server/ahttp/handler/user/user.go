// Package userhandler 提供用户相关的 HTTP 处理器
package userhandler

import (
	"aibuddy/internal/services/auth"
	"aibuddy/pkg/ahttp"
	"aibuddy/pkg/config"
	logger "aibuddy/pkg/log"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = func() trace.Tracer {
	return otel.Tracer(config.Instance.Tracer.ServiceName)
}

// Handler 用户相关处理器
type Handler struct {
	AuthServer *auth.Service
}

// New 创建用户处理器实例
func New() *Handler {
	return &Handler{
		AuthServer: auth.New(),
	}
}

// Login 用户登录
func (h *Handler) Login(state *ahttp.State, req *LoginRequest) error {
	// 从 Echo context 获取请求 context，保持追踪链路
	ctx, span := tracer().Start(state.Ctx.Request().Context(), "login")
	defer span.End()

	span.SetAttributes(attribute.String("username", req.Username))
	span.SetAttributes(attribute.String("password", req.Password))

	slog.Info(logger.Authorization, "username", req.Username, "password", req.Password)

	user, err := h.AuthServer.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	return state.Resposne().SetData(LoginResponse{Token: user}).Success()
}
