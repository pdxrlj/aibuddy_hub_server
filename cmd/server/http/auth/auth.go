// Package auth 提供认证功能
package auth

import (
	"aibuddy/pkg/config"
	"aibuddy/pkg/http"
	"context"
	"errors"
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
func RegisterRoutes(base *http.Base) {
	base.POST("/login", func(state *http.State, req *LoginRequest) error {
		ctx, span := tracer().Start(context.Background(), "login")
		defer span.End()
		_ = ctx
		span.SetAttributes(attribute.String("username", req.Username))
		span.SetAttributes(attribute.String("password", req.Password))

		if req.Username != "admin" || req.Password != "123456" {
			return http.NewResponse(state.Ctx).SetStatus(stdh.StatusUnauthorized).
				SetMessage("Invalid username or password").
				Error(errors.New("invalid username or password"))
		}

		return http.NewResponse(state.Ctx).SetData(map[string]string{"token": "1234567890"}).Success()
	})
}
