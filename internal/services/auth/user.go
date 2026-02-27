// Package auth provides user authentication services.
package auth

import (
	"aibuddy/internal/model"
	"aibuddy/internal/repository"
	"aibuddy/internal/services/cache"
	"aibuddy/pkg/config"
	logger "aibuddy/pkg/log"
	"aibuddy/pkg/wechatservice"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

var testPhoneNumber = []string{"18888888888", "18895516550"}
var testCode = "12345"

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

// CheckLoginCode 检验验证码
func (s *Service) CheckLoginCode(phone, code string) error {
	// 测试数据
	if slices.Contains(testPhoneNumber, phone) && code == testCode {
		return nil
	}
	key := fmt.Sprintf("sms:%s", phone)
	result, err := cache.Flash().Get(key)
	if errors.Is(err, redis.Nil) {
		return errors.New("请先发送验证码")
	} else if err != nil {
		return err
	}
	slog.Info(logger.Authorization, "key", key, "code", result.(string))
	if result.(string) != code {
		return errors.New("验证码错误")
	}
	return nil
}

// CheckLoginMiniProgram 验证微信小程序登录参数
func (s *Service) CheckLoginMiniProgram(code, encryptedData, iv string, userInfo *model.User) (user *model.User, err error) {
	// 获取微信小程序实例
	miniprogram, err := wechatservice.GetWechatMiniProgram()
	if err != nil {
		slog.Error(logger.Authorization, "msg", "Failed to get WeChat mini program instance", "error", err)
		return nil, errors.New("failed to get WeChat mini program instance")
	}

	// 调用微信登录接口获取 session 信息
	session, err := miniprogram.GetAuth().Code2Session(code)
	if err != nil {
		slog.Error(logger.Authorization, "msg", "Failed to exchange WeChat code for session", "error", err)
		return nil, errors.New("登录参数不合法")
	}

	// 获取用户手机号
	plainData, err := miniprogram.GetEncryptor().Decrypt(session.SessionKey, encryptedData, iv)
	if err != nil {
		slog.Error(logger.Authorization, "msg", "Failed to decrypt WeChat encrypted data", "error", err)
		return nil, errors.New("参数异常")
	}

	userInfo.OpenID = session.OpenID
	userInfo.Phone = plainData.PhoneNumber
	userInfo.Nickname = plainData.NickName
	userInfo.Avatar = plainData.AvatarURL

	return userInfo, nil
}
