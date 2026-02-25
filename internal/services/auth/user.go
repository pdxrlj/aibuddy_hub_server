// Package auth provides user authentication services.
package auth

import (
	"aibuddy/pkg/config"
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Service 用户认证服务
type Service struct{}

var tracer = func() trace.Tracer {
	return otel.Tracer(config.Instance.Tracer.ServiceName)
}

// New 创建用户认证服务实例
func New() *Service {
	return &Service{}
}

// GetUserByUsername 根据用户名获取用户信息
func (s *Service) GetUserByUsername(ctx context.Context, username string) (string, error) {
	_, span := tracer().Start(ctx, "GetUserByUsername")
	defer span.End()

	span.SetAttributes(attribute.String("username", username))

	return "123456", nil
}
