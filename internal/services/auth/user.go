// Package auth provides user authentication services.
package auth

import (
	management "aibuddy/aiframe/managemet"
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"aibuddy/internal/repository"
	"aibuddy/internal/services/cache"
	"aibuddy/internal/services/device"
	"aibuddy/pkg/config"
	"aibuddy/pkg/flash"
	"aibuddy/pkg/helpers"
	logger "aibuddy/pkg/log"
	"aibuddy/pkg/sms"
	"aibuddy/pkg/wechatservice"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

var (
	testPhoneNumber = []string{"18888888888", "18895516550"}
	testCode        = "123456"

	smsCacheKey     = "sms:%s"
	smsSendCountKey = "sms_count:%s"
)

// Service 用户认证服务
type Service struct {
	UserRepo         *repository.UserRepo
	DeviceInfoRepo   *repository.DeviceInfoRepo
	DeviceRepo       *repository.DeviceRepo
	BindDeviceSnRepo *repository.BindDeviceSnRepo

	sms   *sms.AliyunSMS
	cache flash.Flash

	deviceService *device.Service
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
		UserRepo:         repository.New(),
		DeviceInfoRepo:   repository.NewDeviceInfoRepo(),
		BindDeviceSnRepo: repository.NewBindDeviceSnRepo(),
		sms:              sms,
		cache:            cache.Flash(),
		deviceService:    device.NewService(),
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
func (s *Service) UpsertUser(ctx context.Context, user *model.User, source string) error {
	ctx, span := tracer().Start(ctx, "UpsertUser")
	defer span.End()

	if user.Nickname == "" && source == "phone" {
		user.Nickname = "默认用户"
	}

	return s.UserRepo.Upsert(ctx, user)
}

// CheckLoginCode 检验验证码
func (s *Service) CheckLoginCode(phone, code string) error {
	if slices.Contains(testPhoneNumber, phone) && code == testCode {
		return nil
	}
	cachekey := fmt.Sprintf("sms:%s", phone)
	result, err := s.cache.Get(cachekey)

	if err != nil && err.Error() != redis.Nil.Error() {
		return err
	} else if err != nil && err.Error() == redis.Nil.Error() {
		return errors.New("验证码不存在或已过期，请重新获取")
	}

	if result.(string) != code {
		return errors.New("验证码错误")
	}

	if err := s.cache.Delete(cachekey); err != nil {
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
func (s *Service) SendPhoneCode(ctx context.Context, phone string) (string, error) {
	_, span := tracer().Start(ctx, "SendPhoneCode")
	defer span.End()

	maxCount := config.Instance.App.MsgSendCount
	cacheKey := fmt.Sprintf(smsCacheKey, phone)
	sendCountKey := fmt.Sprintf(smsSendCountKey, phone)

	if s.cache.Exists(cacheKey) {
		span.RecordError(errors.New("验证码已发送，请稍后重试"))
		return "", errors.New("验证码已发送，请稍后重试")
	}

	// 生成短信验证码
	code := helpers.GenerateNumber(6)
	if !slices.Contains(testPhoneNumber, phone) {
		// 先检查发送次数
		if _, err := s.LimitTaskTimes(sendCountKey, maxCount, 24*time.Hour); err != nil {
			span.RecordError(err)
			return "", err
		}

		// 发送验证码
		_, err := s.sms.SendSMS(phone, code)
		if err != nil {
			span.RecordError(err)
			return "", err
		}
	} else {
		code = testCode
	}

	if err := s.cache.Set(cacheKey, code, 5*time.Minute); err != nil {
		span.RecordError(err)
		return "", err
	}

	return code, nil
}

// LimitTaskTimes 限制规定时间任务次数，返回当前计数
func (s *Service) LimitTaskTimes(key string, times int, ttl time.Duration) (int, error) {
	if times < 1 {
		times = 1
	}
	count, err := s.cache.Incr(key, ttl)
	if err != nil {
		return 0, err
	}

	if count > int64(times) {
		return int(count), errors.New("任务次数达到最大限制")
	}

	return int(count), nil
}

// CompleteProfile 完善设备信息
func (s *Service) CompleteProfile(ctx context.Context, uid int64, boardType string, d *model.DeviceInfo) error {
	ctx, span := tracer().Start(ctx, "CompleteProfile")
	defer span.End()

	return query.Q.Transaction(func(tx *query.Query) error {
		if err := s.DeviceInfoRepo.UpsertProfile(ctx, d, tx); err != nil {
			return err
		}
		if err := s.DeviceRepo.FirstAddDevice(ctx, d.DeviceID, uid, tx); err != nil {
			return err
		}

		iccid, version, err := s.deviceService.FromCacheGetDeviceInfo(d.DeviceID)
		if err != nil {
			return err
		}

		if err := s.DeviceRepo.ChangeDeviceInfo(ctx, d.DeviceID, iccid, boardType, version, tx); err != nil {
			return err
		}

		user, err := s.UserRepo.FindUserByUserID(uid)
		if err != nil {
			return err
		}

		// 如果表里面没有这个Device那可能就是非法的Device
		sn, err := s.BindDeviceSnRepo.GetDeviceSnByDeviceID(ctx, d.DeviceID)
		if err != nil {
			return err
		}

		mMgmt := management.Mgmt{
			Type:   management.MgmtTypeBound,
			User:   user.Nickname,
			Avatar: user.Avatar,
			Sn:     sn,
		}

		if err := mMgmt.SendBoundToDevice(d.DeviceID); err != nil {
			return err
		}

		return nil
	})
}
