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
type Service struct {
	UserRepo *repository.UserRepo
}

var tracer = func() trace.Tracer {
	return otel.Tracer(config.Instance.Tracer.ServiceName)
}

// New 创建用户认证服务实例
func New() *Service {
	return &Service{
		UserRepo: repository.New(),
	}
}

// GetUserByPhone 根据手机号获取用户信息
func (s *Service) GetUserByPhone(ctx context.Context, phone string) (*model.User, error) {
	_, span := tracer().Start(ctx, "GetUserByPhone")
	defer span.End()

	span.SetAttributes(attribute.String("phone", phone))

	user, err := s.UserRepo.FindUserInfoByPhone(phone)
	if err != nil && err.Error() != gorm.ErrRecordNotFound.Error() {
		return nil, err
	}

	return user, nil
}

// CreateUser 创建用户
func (s *Service) CreateUser(ctx context.Context, user *model.User) (int64, error) {
	_, span := tracer().Start(ctx, "CreateUser")
	defer span.End()

	return s.UserRepo.CreateUser(user)
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

	_, err := s.UserRepo.UpdateUser(uid, user)
	if err != nil {
		return err
	}
	return nil
}

// UpsertUser 插入或更新用户
func (s *Service) UpsertUser(ctx context.Context, user *model.User) error {
	ctx, span := tracer().Start(ctx, "UpsertUser")
	defer span.End()

	return s.UserRepo.Upsert(ctx, user)
}
