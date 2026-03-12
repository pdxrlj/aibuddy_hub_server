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
	"slices"
	"strconv"
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
)

// Service 用户认证服务
type Service struct {
	UserRepo         *repository.UserRepo
	DeviceInfoRepo   *repository.DeviceInfoRepo
	DeviceRepo       *repository.DeviceRepo
	BindDeviceSnRepo *repository.BindDeviceSnRepo
	DeviceMsgRepo    *repository.DeviceMessageRepo
	GrowthReportRepo *repository.GrowthReportRepo

	sms   *sms.AliyunSMS
	cache flash.Flash

	deviceService *device.Service

	growthReportService *agent.GrowthReport
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

		growthReportService: agent.NewGroupReport(),
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
func (s *Service) CompleteProfile(ctx context.Context, uid int64, boardType string, relation string, d *model.DeviceInfo) error {
	ctx, span := tracer().Start(ctx, "CompleteProfile")
	defer span.End()

	return query.Q.Transaction(func(tx *query.Query) error {
		if err := s.DeviceInfoRepo.UpsertProfile(ctx, d, tx); err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("device_id", d.DeviceID))
			span.SetAttributes(attribute.String("error", err.Error()))
			return err
		}
		if err := s.DeviceRepo.FirstAddDevice(ctx, d.DeviceID, uid, tx); err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("device_id", d.DeviceID))
			span.SetAttributes(attribute.String("error", err.Error()))
			return err
		}

		simCard, version, err := s.deviceService.FromCacheGetDeviceInfo(d.DeviceID)
		if err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("device_id", d.DeviceID))
			span.SetAttributes(attribute.String("error", err.Error()))
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
			return err
		}

		user, err := s.UserRepo.FindUserByUserID(uid)
		if err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("device_id", d.DeviceID))
			span.SetAttributes(attribute.String("error", err.Error()))
			return err
		}

		// 如果表里面没有这个Device那可能就是非法的Device
		sn, err := s.BindDeviceSnRepo.GetDeviceSnByDeviceID(ctx, d.DeviceID)
		if err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("device_id", d.DeviceID))
			span.SetAttributes(attribute.String("error", err.Error()))
			return errors.New("当前设备没有找到合法的SN编码，请检查设备")
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
		return err
	}

	contact := user.Nickname
	if contact == "" {
		contact = "家长"
	}

	if err := management.SendLost(deviceID, contact, user.Phone); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		return err
	}

	if err := s.DeviceRepo.SetDeviceStatus(deviceID, model.DeviceStatusLost); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
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
		return err
	}

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
		return false, err
	}

	return len(deviceInfo) == 0, nil
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
		return err
	}

	data.MsgID = helpers.GenerateNumber(10)
	data.FromDeviceID = fmt.Sprintf("%d", uid)
	data.FromUsername = info.Relation
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

// GetMessage 获取指定留言
func (s *Service) GetMessage(ctx context.Context, uid int64, deviceID string, page int, pageSize int) ([]*model.DeviceMessage, int64, error) {
	_, span := tracer().Start(ctx, "CreateMessage")
	defer span.End()

	if !s.DeviceRepo.CheckDeviceAuth(ctx, uid, deviceID) {
		err := errors.New("设备ID参数错误")
		span.RecordError(err)
		span.SetAttributes(attribute.Int64("uid", uid), attribute.String("device_id", deviceID))
		return nil, 0, err
	}

	return s.DeviceMsgRepo.GetMessageBetweenUser(ctx, deviceID, strconv.Itoa(int(uid)), page, pageSize)
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
		StartTime:            startTime,
		EndTime:              endTime,
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

// mustMarshal 将任意类型序列化为 JSON，失败时返回 nil
func mustMarshal(v any) []byte {
	data, _ := json.Marshal(v)
	return data
}

// GetGrowthReportList 获取用户成长报告列表
func (s *Service) GetGrowthReportList(ctx context.Context, deviceID string, page, pageSize int) ([]*GrowthReportResponse, int64, error) {
	_, span := tracer().Start(ctx, "GetGrowthReportList")
	defer span.End()

	reports, total, err := s.GrowthReportRepo.GetListByDeviceID(ctx, deviceID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	formatReports := make([]*GrowthReportResponse, 0, len(reports))

	for i := range reports {
		growthReport, err := s.FormatGrowthReport(ctx, reports[i])
		if err != nil {
			return nil, 0, err
		}
		formatReports = append(formatReports, growthReport)
	}

	return formatReports, total, nil
}
