// Package aiuser provides user authentication services.
package aiuser

import (
	management "aibuddy/aiframe/managemet"
	"aibuddy/aiframe/message"
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"aibuddy/internal/repository"
	"aibuddy/internal/services/agent"
	"aibuddy/internal/services/cache"
	"aibuddy/internal/services/device"
	"aibuddy/internal/services/websocket"
	"aibuddy/pkg/config"
	"aibuddy/pkg/flash"
	"aibuddy/pkg/helpers"
	logger "aibuddy/pkg/log"
	"aibuddy/pkg/sms"
	"aibuddy/pkg/wechatservice"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/cast"
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

	// DefaultDeviceActivateTime 默认设备激活时间 365天
	DefaultDeviceActivateTime = 365 * 24 * 60 * 60 * time.Second
)

// getDefaultAgentName 获取默认 Agent 名称
func getDefaultAgentName() string {
	if config.Instance.Baidu != nil && config.Instance.Baidu.Agent != nil && config.Instance.Baidu.Agent.Default != "" {
		return config.Instance.Baidu.Agent.Default
	}
	return "阳光元气型"
}

// Service 用户认证服务
type Service struct {
	UserRepo           *repository.UserRepo
	DeviceInfoRepo     *repository.DeviceInfoRepo
	DeviceRepo         *repository.DeviceRepo
	BindDeviceSnRepo   *repository.BindDeviceSnRepo
	DeviceMsgRepo      *repository.DeviceMessageRepo
	GrowthReportRepo   *repository.GrowthReportRepo
	EmotionRepo        *repository.EmotionRepo
	DeviceActivateRepo *repository.BuddyDeviceActivateRepo
	MemberShopRepo     *repository.MemberShopRepository

	sms   *sms.AliyunSMS
	cache flash.Flash

	deviceService *device.Service

	growthReportService *agent.GrowthReport

	AfterCompleteProfileHook []AfterCompleteProfileHook
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
		UserRepo:         repository.NewUserRepo(),
		DeviceInfoRepo:   repository.NewDeviceInfoRepo(),
		BindDeviceSnRepo: repository.NewBindDeviceSnRepo(),
		sms:              sms,
		cache:            cache.Flash(),
		deviceService:    device.NewService(),
		DeviceMsgRepo:    repository.NewDeviceMessageRepo(),
		GrowthReportRepo: repository.NewGrowthReportRepo(),
		EmotionRepo:      repository.NewEmotionRepo(),
		DeviceRepo:       repository.NewDeviceRepo(),
		MemberShopRepo:   repository.NewMemberShopRepository(),

		growthReportService: agent.NewGroupReport(),
		AfterCompleteProfileHook: []AfterCompleteProfileHook{
			AfterCompleteProfileSendChildInfo,
		},
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

// GetUserInfoByUID 获取用户详细信息
func (s *Service) GetUserInfoByUID(ctx context.Context, uid int64) (*model.User, error) {
	ctx, span := tracer().Start(ctx, "CreateUser")
	defer span.End()

	return s.UserRepo.GetUserByUID(ctx, uid)
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

// UpdateUserInfo 更新用户信息
func (s *Service) UpdateUserInfo(ctx context.Context, uid int64, data *model.User) error {
	_, span := tracer().Start(ctx, "UpdateUserInfo")
	defer span.End()

	if data.Avatar != "" {
		re := regexp.MustCompile(`(?i)\.(jpg|jpeg|png)$`)
		if !re.MatchString(data.Avatar) {
			span.SetAttributes(attribute.String("avatar", data.Avatar))
			return errors.New("图像参数异常")
		}
	}

	_, err := s.UserRepo.UpdateUser(uid, &model.User{
		Nickname: data.Nickname,
		Avatar:   data.Avatar,
		Phone:    data.Phone,
		Email:    data.Email,
		Birthday: data.Birthday,
		Username: data.Username,
		Gender:   data.Gender,
	})
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

	if err := helpers.ValidateMobile(phone); err != nil {
		return errors.New("手机号码格式错误")
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
func (s *Service) CheckLoginMiniProgram(ctx context.Context, code, encryptedData, iv string, userInfo *model.User) (user *model.User, err error) {
	_, span := tracer().Start(ctx, "CheckLoginMiniProgram")
	defer span.End()
	// 获取微信小程序实例
	miniprogram, err := wechatservice.GetWechatMiniProgram()
	if err != nil {
		slog.Error(logger.Authorization, "msg", "Failed to get WeChat mini program instance", "error", err)
		return nil, errors.New("failed to get WeChat mini program instance")
	}

	// 调用微信登录接口获取 session 信息
	session, err := miniprogram.GetAuth().Code2Session(code)
	if err != nil {
		span.SetAttributes(attribute.String("err", err.Error()))
		span.RecordError(err)
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
//
//nolint:cyclop
func (s *Service) CompleteProfile(ctx context.Context, uid int64, boardType, relation string, d *model.DeviceInfo) error {
	ctx, span := tracer().Start(ctx, "CompleteProfile")
	defer span.End()

	return query.Q.Transaction(func(tx *query.Query) error {
		slog.Info("[CompleteProfile] UpsertProfile")
		if err := s.DeviceInfoRepo.UpsertProfile(ctx, d, tx); err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("device_id", d.DeviceID))
			span.SetAttributes(attribute.String("error", err.Error()))
			slog.Info("[CompleteProfile] UpsertProfile", "error", err.Error())
			return err
		}
		if err := s.DeviceRepo.FirstAddDevice(ctx, d.DeviceID, uid, tx); err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("device_id", d.DeviceID))
			span.SetAttributes(attribute.String("error", err.Error()))
			slog.Info("[CompleteProfile] FirstAddDevice", "error", err.Error())
			return err
		}

		simCard, version, err := s.deviceService.FromCacheGetDeviceInfo(strings.ToUpper(d.DeviceID))
		if err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("device_id", d.DeviceID))
			span.SetAttributes(attribute.String("error", err.Error()))
			slog.Error("[CompleteProfile] FromCacheGetDeviceInfo", "error", err.Error())
			return err
		}

		if err := s.DeviceRepo.ChangeDeviceInfo(ctx, d.DeviceID, simCard, boardType, version, relation, tx); err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("device_id", d.DeviceID))
			span.SetAttributes(attribute.String("sim_card", simCard))
			span.SetAttributes(attribute.String("board_type", boardType))
			span.SetAttributes(attribute.String("version", version))
			span.SetAttributes(attribute.String("relation", relation))
			span.SetAttributes(attribute.String("error", err.Error()))
			slog.Error("[CompleteProfile] ChangeDeviceInfo", "error", err.Error())
			return err
		}

		user, err := s.UserRepo.FindUserByUserID(uid)
		if err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("device_id", d.DeviceID))
			span.SetAttributes(attribute.String("error", err.Error()))
			slog.Error("[CompleteProfile] FindUserByUserID", "error", err.Error())
			return err
		}
		// 如果表里面没有这个Device那可能就是非法的Device
		sn, err := s.BindDeviceSnRepo.GetDeviceSnByDeviceID(ctx, d.DeviceID)
		if err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("device_id", d.DeviceID))
			span.SetAttributes(attribute.String("error", err.Error()))
			slog.Error("[CompleteProfile] GetDeviceSnByDeviceID", "error", err.Error())
			return errors.New("当前设备没有找到合法的SN编码，请检查设备")
		}

		// 给Device设置默认的Agent
		if err := s.DeviceRepo.SetDeviceAgent(d.DeviceID, getDefaultAgentName(), tx); err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("device_id", d.DeviceID))
			span.SetAttributes(attribute.String("error", err.Error()))
			slog.Info("[CompleteProfile] SetDeviceAgent", "error", err.Error())
			return err
		}

		if err := s.CheckAndActivateDevice(ctx, d.DeviceID, tx); err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("device_id", d.DeviceID))
			span.SetAttributes(attribute.String("error", err.Error()))
			slog.Info("[CompleteProfile] CheckAndActivateDevice", "error", err.Error())
			return err
		}

		mMgmt := management.Mgmt{
			Type:   management.MgmtTypeBound,
			User:   user.Nickname,
			Avatar: user.Avatar,
			Sn:     sn,
		}
		if err := mMgmt.SendBoundToDevice(d.DeviceID); err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("device_id", d.DeviceID))
			span.SetAttributes(attribute.String("error", err.Error()))
			slog.Info("[CompleteProfile] SendBoundToDevice", "error", err.Error())
			return err
		}

		return nil
	})
}

// Lost 挂失设备
func (s *Service) Lost(ctx context.Context, uid int64, deviceID string) error {
	_, span := tracer().Start(ctx, "Lost")
	defer span.End()

	if !s.DeviceRepo.IsValidDevice(ctx, deviceID) {
		span.RecordError(errors.New("设备不存在"))
		span.SetAttributes(attribute.String("device_id", deviceID))
		return errors.New("设备不存在")
	}

	user, err := s.UserRepo.FindUserByUserID(uid)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.Int64("user_id", uid))
		span.SetAttributes(attribute.String("device_id", deviceID))
		slog.Error("[CompleteProfile] Lost", "error", err.Error())
		return err
	}

	contact := user.Nickname
	if contact == "" {
		contact = "家长"
	}

	if err := management.SendLost(deviceID, contact, user.Phone); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		slog.Error("[CompleteProfile] Unlost", "error", err.Error())
		return err
	}

	if err := s.DeviceRepo.SetDeviceStatus(deviceID, model.DeviceStatusLost); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		slog.Error("[CompleteProfile] Unbind", "error", err.Error())
		return err
	}

	return nil
}

// Unlost 解除挂失设备
func (s *Service) Unlost(ctx context.Context, deviceID string) error {
	_, span := tracer().Start(ctx, "Unlost")
	defer span.End()

	if !s.DeviceRepo.IsValidDevice(ctx, deviceID) {
		span.RecordError(errors.New("设备不存在"))
		span.SetAttributes(attribute.String("device_id", deviceID))
		return errors.New("设备不存在")
	}

	if err := management.SendUnlost(deviceID); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		return err
	}

	if err := s.DeviceRepo.SetDeviceStatus(deviceID, model.DeviceStatusUnknown); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		return err
	}

	return nil
}

// Unbind 解绑设备
func (s *Service) Unbind(ctx context.Context, deviceID string) error {
	_, span := tracer().Start(ctx, "Unbind")
	defer span.End()

	if !s.DeviceRepo.IsValidDevice(ctx, deviceID) {
		span.RecordError(errors.New("设备不存在"))
		span.SetAttributes(attribute.String("device_id", deviceID))
		return errors.New("设备不存在")
	}

	if err := management.SendUnboundToDevice(deviceID); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		slog.Error("[CompleteProfile] HaveDevice", "error", err.Error())
		return err
	}
	slog.Info("[Unbind] SendUnboundToDevice success", "device_id", deviceID)
	if err := s.DeviceRepo.EraseDevice(ctx, deviceID); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		return err
	}

	return nil
}

// HaveDevice 用户是否已经绑定了设备
func (s *Service) HaveDevice(ctx context.Context, uid int64) (bool, error) {
	_, span := tracer().Start(ctx, "HaveDevice")
	defer span.End()

	deviceInfo, err := s.DeviceRepo.GetUserDeviceList(ctx, uid)

	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.Int64("user_id", uid))
		span.SetAttributes(attribute.String("error", err.Error()))
		return false, err
	}

	return len(deviceInfo) > 0, nil
}

// UserDeviceList 设备列表
func (s *Service) UserDeviceList(ctx context.Context, uid int64) ([]*model.Device, error) {
	_, span := tracer().Start(ctx, "UserDeviceList")
	defer span.End()

	devices, err := s.DeviceRepo.GetUserDeviceList(ctx, uid)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.Int64("user_id", uid))
		span.SetAttributes(attribute.String("error", err.Error()))
		slog.Error("[CompleteProfile] UserDeviceList", "error", err.Error())
		return nil, err
	}

	return devices, nil
}

// CreateMessage 用户留言
func (s *Service) CreateMessage(ctx context.Context, uid int64, data *model.DeviceMessage) error {
	_, span := tracer().Start(ctx, "CreateMessage")
	defer span.End()

	if !s.DeviceRepo.CheckDeviceAuth(ctx, uid, data.ToDeviceID) {
		err := errors.New("无法给未绑定的设备留言")
		span.RecordError(err)
		span.SetAttributes(attribute.Int64("uid", uid), attribute.String("device_id", data.ToDeviceID))
		return err
	}
	info, err := s.DeviceRepo.GetDeviceInfo(ctx, data.ToDeviceID)
	if err != nil || info == nil {
		err := errors.New("设备信息异常")
		span.RecordError(err)
		span.SetAttributes(attribute.Int64("uid", uid), attribute.String("device_id", data.ToDeviceID))
		slog.Error("[CompleteProfile] CreateMessage", "error", err.Error())
		return err
	}

	data.MsgID = helpers.GenerateNumber(10)
	data.FromDeviceID = fmt.Sprintf("%d", uid)

	if err = s.DeviceMsgRepo.CreateDeviceMessage(ctx, data); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", data.ToDeviceID), attribute.Int64("uid", uid))
		return errors.New("创建留言失败")
	}

	// 通过 MQTT 发送消息给设备
	if err := message.SendMessage(strconv.Itoa(int(uid)), info.DeviceInfo.NickName, info.DeviceID, data.MsgID, data.Content, data.Fmt.String(), data.Dur); err != nil {
		span.RecordError(err)
		return errors.New("给设备留言失败")
	}

	return nil
}

// GetMessage 获取指定留言（按对话分组）
func (s *Service) GetMessage(ctx context.Context, uid int64, deviceID string, page int, pageSize int) ([][]*MessageDTO, int64, error) {
	_, span := tracer().Start(ctx, "CreateMessage")
	defer span.End()
	// slog.Info("GetMessage", "uid", uid, "device_id", deviceID, "page", page, "pageSize", pageSize)
	if !s.DeviceRepo.CheckDeviceAuth(ctx, uid, deviceID) {
		err := errors.New("设备ID参数错误")
		span.RecordError(err)
		slog.Error("[GetMessage] CheckDeviceAuth error", "uid", uid, "device_id", deviceID, "error", err.Error())
		span.SetAttributes(attribute.Int64("uid", uid), attribute.String("device_id", deviceID))
		return nil, 0, err
	}

	messages, total, err := s.DeviceMsgRepo.GetMessageFromUser(ctx, deviceID, page, pageSize)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.Int64("uid", uid), attribute.String("device_id", deviceID))
		return nil, 0, err
	}
	dtoMessages := s.ToMessageDTO(messages)
	return dtoMessages, total, nil
}

// AnalysisGrowthReport 分析用户成长报告
func (s *Service) AnalysisGrowthReport(ctx context.Context, uid int64, deviceID string, startTime, endTime time.Time) error {
	_, span := tracer().Start(ctx, "AnalysisGrowthReport")
	defer span.End()

	uidStr := cast.ToString(uid)

	go func() {
		ctx, span := tracer().Start(context.Background(), "AnalysisGrowthReport.RunGrowthReport")
		defer span.End()

		result := &struct {
			err       error
			msg       []byte
			isSuccess bool
		}{}

		defer func() {
			slog.Info("[AnalysisGrowthReport] SendMessage", "device_id", deviceID, "result", result.msg)
			websocket.SendMessage(uidStr, &websocket.GrowthReportFrame{
				Type:     helpers.Cond(result.isSuccess, websocket.FrameTypeGrowthReportSuccess, websocket.FrameTypeGrowthReportFailure),
				DeviceID: deviceID,
				Message:  result.msg,
			})
		}()

		report, err := s.growthReportService.RunGrowthReport(ctx, deviceID, startTime, endTime)
		if err != nil {
			span.RecordError(err)
			slog.Error("[AnalysisGrowthReport] RunGrowthReport failed", "device_id", deviceID, "error", err)
			result.err = err
			return
		}

		if report == nil {
			err := errors.New("report is nil")
			span.RecordError(err)
			slog.Error("[AnalysisGrowthReport] report is nil", "device_id", deviceID)
			result.err = err
			return
		}

		growthReport, err := s.ConvertToGrowthReport(deviceID, startTime, endTime, report)
		if err != nil {
			span.RecordError(err)
			slog.Error("[AnalysisGrowthReport] convertToGrowthReport failed", "device_id", deviceID, "error", err)
			result.err = err
			return
		}

		if err := s.GrowthReportRepo.Create(ctx, growthReport); err != nil {
			span.RecordError(err)
			slog.Error("[AnalysisGrowthReport] Create failed", "device_id", deviceID, "error", err)
			result.err = err
			return
		}

		result.msg = growthReport.MustString()
		result.isSuccess = true
		slog.Info("[AnalysisGrowthReport] success", "device_id", deviceID)
	}()

	return nil
}

// ConvertToGrowthReport 将 StageTwoReport 转换为 model.GrowthReport
func (s *Service) ConvertToGrowthReport(deviceID string, startTime, endTime time.Time, report *agent.StageTwoReport) (*model.GrowthReport, error) {
	statusCards := mustMarshal(report.StatusCards)
	interactionSummary := mustMarshal(report.InteractionSummary)
	socialSummary := mustMarshal(report.SocialSummary)
	memoryCapsuleSummary := mustMarshal(report.MemoryCapsuleSummary)
	childPortrait := mustMarshal(report.ChildPortrait)
	keyMoments := mustMarshal(report.KeyMoments)
	emotionTrend := mustMarshal(report.EmotionTrend)
	audioSummary := mustMarshal(report.AudioSummary)
	pomodoroSummary := mustMarshal(report.PomodoroSummary)
	safetyAlert := mustMarshal(report.SafetyAlert)
	nextWeekSuggestions := mustMarshal(report.NextWeekSuggestions)
	parentScripts := mustMarshal(report.ParentScripts)

	return &model.GrowthReport{
		DeviceID:             deviceID,
		StartTime:            model.LocalTime(startTime),
		EndTime:              model.LocalTime(endTime),
		SummaryText:          report.SummaryText,
		StatusCards:          statusCards,
		InteractionSummary:   interactionSummary,
		SocialSummary:        socialSummary,
		MemoryCapsuleSummary: memoryCapsuleSummary,
		ChildPortrait:        childPortrait,
		KeyMoments:           keyMoments,
		EmotionTrend:         emotionTrend,
		AudioSummary:         audioSummary,
		PomodoroSummary:      pomodoroSummary,
		SafetyAlert:          safetyAlert,
		NextWeekSuggestions:  nextWeekSuggestions,
		ParentScripts:        parentScripts,
		ClosingText:          report.ClosingText,
	}, nil
}

// GetUserDeviceInfo 获取用户设备信息
func (s *Service) GetUserDeviceInfo(ctx context.Context, uid int64, deviceID string) (*model.DeviceInfo, *model.Device, error) {
	ctx, span := tracer().Start(ctx, "GetUserDeviceInfo")
	defer span.End()

	if !s.DeviceRepo.CheckDeviceAuth(ctx, uid, deviceID) {
		span.SetAttributes(attribute.Int64("uid", uid), attribute.String("device_id", deviceID))
		return nil, nil, errors.New("无法获取该设备的信息")
	}

	info, err := s.DeviceInfoRepo.GetUserInfoByDeivceID(ctx, deviceID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.Int64("uid", uid), attribute.String("device_id", deviceID))
		return nil, nil, err
	}

	deviceInfo, err := s.DeviceRepo.GetDeviceInfo(ctx, deviceID)
	if err != nil {
		return nil, nil, err
	}

	return info, deviceInfo, nil
}

// UpdateDeviceInfo 获取用户设备信息
func (s *Service) UpdateDeviceInfo(ctx context.Context, uid int64, data *model.DeviceInfo, relation string) error {
	ctx, span := tracer().Start(ctx, "UpdateDeviceInfo")
	defer span.End()

	if !s.DeviceRepo.CheckDeviceAuth(ctx, uid, data.DeviceID) {
		span.SetAttributes(attribute.Int64("uid", uid), attribute.String("device_id", data.DeviceID))
		return errors.New("无法设置该设备的信息")
	}

	// 执行完成用户信息更新后钩子
	for _, hook := range s.AfterCompleteProfileHook {
		if err := hook(ctx, data.DeviceID); err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("device_id", data.DeviceID))
			return err
		}
	}

	return s.DeviceInfoRepo.UpdateDeviceInfo(ctx, data, uid, relation)
}

// mustMarshal 将任意类型序列化为 JSON，失败时返回 nil
func mustMarshal(v any) []byte {
	data, _ := json.Marshal(v)
	return data
}

// GrowthReportListItem 成长报告列表项
type GrowthReportListItem struct {
	DeviceName string `json:"device_name"`
	*model.GrowthReport
}

// GetGrowthReportList 获取用户成长报告列表
func (s *Service) GetGrowthReportList(ctx context.Context, deviceID string, page, pageSize int) ([]*GrowthReportListItem, int64, error) {
	_, span := tracer().Start(ctx, "GetGrowthReportList")
	defer span.End()

	reports, total, err := s.GrowthReportRepo.GetListByDeviceID(ctx, deviceID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// formatReports := make([]*GrowthReportResponse, 0, len(reports))

	// for i := range reports {
	// 	growthReport, err := s.FormatGrowthReport(ctx, reports[i])
	// 	if err != nil {
	// 		return nil, 0, err
	// 	}
	// 	formatReports = append(formatReports, growthReport)
	// }

	var formatReports []*GrowthReportListItem
	device, err := s.DeviceRepo.GetDeviceInfo(ctx, deviceID)
	if err != nil {
		return nil, 0, err
	}

	if device.DeviceInfo == nil {
		return nil, 0, errors.New("设备信息异常,无法获取设备名称")
	}

	for _, report := range reports {
		formatReports = append(formatReports, &GrowthReportListItem{
			DeviceName:   device.DeviceInfo.NickName,
			GrowthReport: report,
		})
	}

	return formatReports, total, nil
}

// DeleteGrowthReport 删除成长报告
func (s *Service) DeleteGrowthReport(ctx context.Context, reportID string) error {
	_, span := tracer().Start(ctx, "DeleteGrowthReport")
	defer span.End()

	err := s.GrowthReportRepo.Delete(ctx, cast.ToInt64(reportID))
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("report_id", reportID))
		return err
	}

	return nil
}

// ClearUserInfo 清除用户全部数据
func (s *Service) ClearUserInfo(ctx context.Context, uid int64) error {
	ctx, span := tracer().Start(ctx, "ClearUserInfo")
	defer span.End()
	deviceList, err := s.DeviceRepo.GetUserDeviceIsAdmin(ctx, uid)
	if err != nil {
		return err
	}

	deviceIDs := make([]string, len(deviceList))
	for _, v := range deviceList {
		deviceIDs = append(deviceIDs, v.DeviceID)
	}

	if err := s.DeviceRepo.ClearDeviceInfo(ctx, uid, deviceIDs); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.Int64("uid", uid))
		return err
	}

	return nil
}

// MessageDTO 留言响应DTO
type MessageDTO struct {
	ID           int    `json:"id"`
	MsgID        string `json:"msg_id"`
	FromDeviceID string `json:"from_device_id"`
	FromUsername string `json:"from_username"`
	ToDeviceID   string `json:"to_device_id"`
	FromAvatar   string `json:"from_avatar"`
	ToAvatar     string `json:"to_avatar"`
	Content      string `json:"content"`
	Fmt          string `json:"fmt"`
	Dur          int    `json:"dur"`
	Read         bool   `json:"read"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// ToMessageDTO 将 DeviceMessage 列表转换为按对话分组的 MessageDTO 列表
// 返回格式: [][]*MessageDTO，每个子数组代表与同一个聊天对象的所有消息
func (s *Service) ToMessageDTO(messages []*model.DeviceMessage) [][]*MessageDTO {
	// 按对话双方分组，确保 A->B 和 B->A 的消息在同一个组
	groups := make(map[string][]*MessageDTO)

	for _, msg := range messages {
		dto := &MessageDTO{
			ID:           msg.ID,
			MsgID:        msg.MsgID,
			FromDeviceID: msg.FromDeviceID,
			ToDeviceID:   msg.ToDeviceID,
			Content:      msg.Content,
			Fmt:          msg.Fmt.String(),
			Dur:          msg.Dur,
			Read:         msg.Read,
			CreatedAt:    time.Time(msg.CreatedAt).Format("2006-01-02 15:04:05"),
			UpdatedAt:    time.Time(msg.UpdatedAt).Format("2006-01-02 15:04:05"),
		}
		// 从 Device.DeviceInfo 获取头像
		if msg.Device != nil && msg.Device.DeviceInfo != nil {
			dto.FromAvatar = msg.Device.DeviceInfo.Avatar
		}
		// 从 ToDevice.DeviceInfo 获取头像
		if msg.ToDevice != nil && msg.ToDevice.DeviceInfo != nil {
			dto.ToAvatar = msg.ToDevice.DeviceInfo.Avatar
		}

		// 生成分组key：将两个deviceID排序后拼接，确保双向对话在同一组
		key := makeConversationKey(msg.FromDeviceID, msg.ToDeviceID)
		groups[key] = append(groups[key], dto)
	}

	// 将 map 转换为二维数组
	result := make([][]*MessageDTO, 0, len(groups))
	for _, group := range groups {
		result = append(result, group)
	}

	return result
}

// makeConversationKey 生成分组key，确保 A-B 和 B-A 的对话使用相同的key
func makeConversationKey(id1, id2 string) string {
	if id1 < id2 {
		return id1 + ":" + id2
	}
	return id2 + ":" + id1
}

// UnreadCountResponse 未读数量响应
type UnreadCountResponse struct {
	MessageCount int64 `json:"message_count"` // 消息未读数量
	EmotionCount int64 `json:"emotion_count"` // 情绪预警未读数量

	LastUnreadMsg *UnreadCountMsgItem `json:"last_unread_msg"` // 最后一条未读消息
	LastEmotion   *UnreadCountMsgItem `json:"last_emotion"`    // 最后一条未读情绪预警
}

// UnreadCountMsgItem 未读消息项
type UnreadCountMsgItem struct {
	Content   string `json:"content"`    // 消息内容
	CreatedAt string `json:"created_at"` // 消息创建时间
}

// GetUnreadMessageCount 获取未读消息数量
func (s *Service) GetUnreadMessageCount(ctx context.Context, uid int64, deviceID string) (*UnreadCountResponse, error) {
	_, span := tracer().Start(ctx, "GetUnreadMessageCount")
	defer span.End()

	response := &UnreadCountResponse{}

	// 对话消息未读数量
	msgCount, lastUnreadMsg, err := s.DeviceMsgRepo.GetUnreadMessageCount(ctx, uid, deviceID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.Int64("uid", uid), attribute.String("device_id", deviceID))
		return nil, err
	}
	response.MessageCount = msgCount
	if lastUnreadMsg != nil {
		response.LastUnreadMsg = &UnreadCountMsgItem{
			Content:   lastUnreadMsg.Content,
			CreatedAt: lastUnreadMsg.CreatedAt.Format(time.DateTime),
		}
	}

	// 情绪预警未读数量
	emotionCount, lastEmotion, err := s.EmotionRepo.GetUnreadCount(ctx, deviceID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		slog.Error("获取情绪预警未读数量失败", "error", err, "device_id", deviceID)
		emotionCount = 0
	}
	response.EmotionCount = emotionCount
	if lastEmotion != nil {
		// 解析WarningTypes JSON数组
		var warningTypes []string
		if err := json.Unmarshal(lastEmotion.WarningTypes, &warningTypes); err != nil {
			// 如果解析失败，使用原始字符串
			warningTypes = []string{lastEmotion.WarningTypes.String()}
		}
		content := strings.Join(warningTypes, ", ")
		response.LastEmotion = &UnreadCountMsgItem{
			Content:   content,
			CreatedAt: lastEmotion.CreatedAt.Format(time.DateTime),
		}
	}

	return response, nil
}

// MyInfoResponse 我的页面信息响应结构体
type MyInfoResponse struct {
	DeviceName        string `json:"device_name"`
	DeviceAvatar      string `json:"device_avatar"`
	DeviceID          string `json:"device_id"`
	Sex               string `json:"sex"`
	FriendCount       int64  `json:"friend_count"`
	FamilyMemberCount int64  `json:"family_member_count"`

	// TODO 会员信息需要根据实际情况返回
	MembershipInfo any `json:"membership_info"`
}

// GetMyInfo 获取我的页面信息：当前设备的信息、 好友数、家庭成员数、会员信息
func (s *Service) GetMyInfo(ctx context.Context, uid int64, deviceID string) (*MyInfoResponse, error) {
	_, span := tracer().Start(ctx, "GetMyInfo")
	defer span.End()

	info, err := s.DeviceRepo.GetDeviceInfo(ctx, deviceID)
	if err != nil {
		return nil, err
	}
	_ = uid
	// 获取好友数量
	_, friendCount, err := s.deviceService.GetFriends(ctx, deviceID, 1, 1)
	if err != nil {
		return nil, err
	}
	// 获取VIP信息
	var memberInfo = map[string]any{
		"member_type": "",
		"member_name": "",
		"expire_time": "",
		"role_num":    0,
	}
	if info.ExpireTime.Unix() > 0 {
		memberInfo["expire_time"] = info.ExpireTime.Format(time.DateOnly)
	}
	member, err := s.MemberShopRepo.GetMemberByDeviceID(ctx, deviceID)
	if err != nil {
		return nil, err
	}
	if member != nil {
		memberInfo["member_type"] = member.Goods[0].GoodsInfo.Name
		memberInfo["member_name"] = member.Goods[0].GoodsInfo.Name
	}
	memberInfo["role_num"] = s.MemberShopRepo.GetMemberRoleNum(ctx, deviceID)

	return &MyInfoResponse{
		DeviceName:        info.DeviceInfo.NickName,
		DeviceAvatar:      info.DeviceInfo.Avatar,
		FriendCount:       friendCount,
		DeviceID:          deviceID,
		Sex:               info.DeviceInfo.Gender,
		FamilyMemberCount: 1,
		MembershipInfo:    memberInfo,
	}, nil
}

// MarkMessageRead 标记消息已读
func (s *Service) MarkMessageRead(ctx context.Context, uid int64, messageIDs []string) error {
	_, span := tracer().Start(ctx, "MarkMessageRead")
	defer span.End()

	if len(messageIDs) == 0 {
		return nil
	}

	err := s.DeviceMsgRepo.BatchMessageRead(ctx, messageIDs)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.Int64("uid", uid), attribute.StringSlice("message_ids", messageIDs))
		return err
	}

	return nil
}

// GetConvMessageList 获取两个设备之间的对话消息列表
func (s *Service) GetConvMessageList(ctx context.Context, uid int64, deviceID, targetDeviceID string, page, pageSize int) ([]*device.MessageDTO, int64, error) {
	_, span := tracer().Start(ctx, "GetConvMessageList")
	defer span.End()

	// 如果为空那就是查询与父母的对话
	if deviceID == "" {
		deviceID = cast.ToString(uid)
	}

	devices, err := s.DeviceRepo.GetUserDeviceList(ctx, uid)
	if err != nil {
		span.RecordError(err)
		return nil, 0, err
	}

	hasPermission := false
	for _, device := range devices {
		if device.DeviceID == deviceID || device.DeviceID == targetDeviceID {
			hasPermission = true
			break
		}
	}

	if !hasPermission {
		err := errors.New("无权限查看该对话")
		span.RecordError(err)
		span.SetAttributes(
			attribute.Int64("uid", uid),
			attribute.String("device_id", deviceID),
			attribute.String("target_device_id", targetDeviceID),
		)
		return nil, 0, err
	}

	messages, total, err := s.DeviceMsgRepo.GetConvMessageList(ctx, deviceID, targetDeviceID, page, pageSize)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(
			attribute.String("device_id", deviceID),
			attribute.String("target_device_id", targetDeviceID),
			attribute.Int("page", page),
			attribute.Int("page_size", pageSize),
		)
		return nil, 0, err
	}

	// 转换为 MessageDTO 格式
	dtoMessages := s.toMessageDTO(messages)

	return dtoMessages, total, nil
}

// toMessageDTO 将 DeviceMessage 列表转换为 MessageDTO 列表
func (s *Service) toMessageDTO(messages []*model.DeviceMessage) []*device.MessageDTO {
	result := make([]*device.MessageDTO, 0, len(messages))
	for _, msg := range messages {
		dto := &device.MessageDTO{
			ID:           msg.ID,
			MsgID:        msg.MsgID,
			FromDeviceID: msg.FromDeviceID,
			ToDeviceID:   msg.ToDeviceID,
			Content:      msg.Content,
			Fmt:          msg.Fmt.String(),
			Dur:          msg.Dur,
			Read:         msg.Read,
			CreatedAt:    time.Time(msg.CreatedAt).Format(time.DateTime),
			UpdatedAt:    time.Time(msg.UpdatedAt).Format(time.DateTime),
		}
		// 从 Device.DeviceInfo 获取头像和用户名
		if msg.Device != nil && msg.Device.DeviceInfo != nil {
			dto.FromAvatar = msg.Device.DeviceInfo.Avatar
			dto.FromUsername = msg.Device.DeviceInfo.NickName
		}
		// 从 ToDevice.DeviceInfo 获取头像和用户名
		if msg.ToDevice != nil && msg.ToDevice.DeviceInfo != nil {
			dto.ToAvatar = msg.ToDevice.DeviceInfo.Avatar
			dto.ToUsername = msg.ToDevice.DeviceInfo.NickName
		}
		result = append(result, dto)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt > result[j].CreatedAt
	})

	return result
}

// CheckAndActivateDevice 判断设备首次是否已经激活了，如果没有激活，buddy_device修改vip过期时间，默认给一年
func (s *Service) CheckAndActivateDevice(ctx context.Context, deviceID string, tx ...*query.Query) error {
	_, span := tracer().Start(ctx, "CheckAndActivateDevice")
	defer span.End()

	if !s.DeviceActivateRepo.IsActivatedDevice(deviceID) {
		if len(tx) > 0 {
			// 使用传入的事务
			if err := s.DeviceActivateRepo.CreateDeviceActivate(deviceID, tx[0]); err != nil {
				return err
			}

			if err := s.DeviceRepo.SetDeviceVipExpireTime(ctx, deviceID, DefaultDeviceActivateTime, tx[0]); err != nil {
				return err
			}

			return nil
		}

		return s.DeviceActivateRepo.Transaction(func(tx *query.Query) error {
			if err := s.DeviceActivateRepo.CreateDeviceActivate(deviceID, tx); err != nil {
				return err
			}

			if err := s.DeviceRepo.SetDeviceVipExpireTime(ctx, deviceID, DefaultDeviceActivateTime, tx); err != nil {
				return err
			}

			return nil
		})
	}

	return nil
}
