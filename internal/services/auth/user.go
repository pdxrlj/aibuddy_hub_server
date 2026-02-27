// Package auth provides user authentication services.
package auth

import (
	"aibuddy/internal/model"
	"aibuddy/internal/repository"
	"aibuddy/pkg/config"
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
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

// GetParentByPhone 根据手机号获取用户信息
func (s *Service) GetParentByPhone(ctx context.Context, phone string) (*model.User, error) {
	_, span := tracer().Start(ctx, "GetParentByPhone")
	defer span.End()

	span.SetAttributes(attribute.String("phone", phone))

	userrepo := &repository.UserRepo{}
	user, err := userrepo.FindUserInfoByPhone(phone)
	if err != nil && err.Error() != gorm.ErrRecordNotFound.Error() {
		return nil, err
	}

	return user, nil
}

// CreateUser 创建用户
func (s *Service) CreateUser(ctx context.Context, user *model.User) (int64, error) {
	_, span := tracer().Start(ctx, "CreateUser")
	defer span.End()

	userrepo := &repository.UserRepo{}
	return userrepo.CreateUser(user)
}

// UpdateUser 更新用户信息
func (s *Service) UpdateUser(ctx context.Context, uid int64, user *model.User, oldUser *model.User) error {
	_, span := tracer().Start(ctx, "UpdateUser")
	defer span.End()

	if user.OpenID == "" {
		user.OpenID = oldUser.OpenID
	}
	if user.Nickname == "" {
		user.Nickname = oldUser.Nickname
	}
	if user.Avatar == "" {
		user.Avatar = oldUser.Avatar
	}

	userrepo := &repository.UserRepo{}
	_, err := userrepo.UpdateUser(uid, user)
	if err != nil {
		return err
	}
	return nil
}
