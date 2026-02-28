// Package auth provides user authentication services.
package auth

import (
	"aibuddy/internal/model"
	"aibuddy/internal/repository"
	"aibuddy/internal/services/cache"
	"aibuddy/pkg/config"
	"aibuddy/pkg/helpers"
	logger "aibuddy/pkg/log"
	"aibuddy/pkg/sms"
	"aibuddy/pkg/wechatservice"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"time"

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
	sms      *sms.AliyunSMS
}

var tracer = func() trace.Tracer {
	return otel.Tracer(config.Instance.Tracer.ServiceName)
}

// New 创建用户认证服务实例
func New() *Service {
	smsConfig := config.Instance.Aliyun.Sms
	sms, err := sms.NewAliyunSMS(
		sms.WithAccessKeyID(config.Instance.Aliyun.Ak),
		sms.WithAccessKeySecret(config.Instance.Aliyun.Sk),
		sms.WithSignName(smsConfig.SignName),
		sms.WithTemplateCode(smsConfig.TemplateCode),
	)

	if err != nil {
		panic(err)
	}

	return &Service{
		UserRepo: repository.New(),
		sms:      sms,
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
	if err := helpers.ValidateMobile(phone); err != nil {
		return err
	}
	// 测试数据
	if slices.Contains(testPhoneNumber, phone) && code == testCode {
		return nil
	}
	cachekey := fmt.Sprintf("sms:%s", phone)
	result, err := cache.Flash().Get(cachekey)

	if err != nil && err.Error() != redis.Nil.Error() {
		return err
	} else if err != nil && err.Error() == redis.Nil.Error() {
		return errors.New("请先发送验证码")
	}

	if result.(string) != code {
		return errors.New("验证码错误")
	}

	if err := cache.Flash().Delete(cachekey); err != nil {
		return err
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

// SendPhoneCode 发送手机验证码
func (s *Service) SendPhoneCode(phone string) (string, error) {
	if err := helpers.ValidateMobile(phone); err != nil {
		return "", err
	}
	cr := cache.Flash()
	maxCount := config.Instance.App.MsgSendCount
	cacheKey := fmt.Sprintf("sms:%s", phone)
	sendKey := fmt.Sprintf("sms_count:%s", phone)

	if cr.Exists(cacheKey) {
		return "", errors.New("验证码已发送，请稍后重试")
	}

	num := 0
	code := helpers.GenerateNumber(5)
	if !slices.Contains(testPhoneNumber, phone) {
		val, _ := cr.Get(sendKey)
		num, _ = strconv.Atoi(val.(string))

		slog.Error("info", "send_key", sendKey, "num", num, "max", maxCount)
		if num >= maxCount {
			return "", errors.New("发送验证码次数达到最大限制，请稍后重试")
		}

		//  发送验证码
		_, err := s.sms.SendSMS(phone, code)
		if err != nil {
			return "", err
		}
		if num == 0 {
			if err := cr.Set(sendKey, num+1, 24*time.Hour); err != nil {
				return "", errors.New("发送验证码失败")
			}
		} else {
			if err := cr.Set(sendKey, num+1); err != nil {
				return "", errors.New("发送验证码失败")
			}
		}
	} else {
		code = testCode
	}

	if err := cr.Set(cacheKey, code, 5*time.Minute); err != nil {
		return "", err
	}

	return code, nil
}
