// Package auth 提供认证功能
package auth

import (
	"aibuddy/pkg/ahttp"
	"aibuddy/pkg/config"
	logger "aibuddy/pkg/log"
	"context"
	"errors"
	"log/slog"
	stdh "net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = func() trace.Tracer {
	return otel.Tracer(config.Instance.Tracer.ServiceName)
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// RegisterRoutes 注册认证路由
func RegisterRoutes(base *ahttp.Base) {
	base.POST("/login", func(state *ahttp.State, req *LoginRequest) error {
		ctx, span := tracer().Start(context.Background(), "login")
		defer span.End()
		_ = ctx
		span.SetAttributes(attribute.String("username", req.Username))
		span.SetAttributes(attribute.String("password", req.Password))

		slog.Info(logger.Authorization, "username", req.Username, "password", req.Password)

		if req.Username != "admin" || req.Password != "123456" {
			return ahttp.NewResponse(state.Ctx).SetStatus(stdh.StatusUnauthorized).
				SetMessage("Invalid username or password").
				Error(errors.New("invalid username or password"))
		}

		return ahttp.NewResponse(state.Ctx).SetData(map[string]string{"token": "1234567890"}).Success()
	})
}
