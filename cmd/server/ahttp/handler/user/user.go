// Package userhandler 提供用户相关的 HTTP 处理器
package userhandler

import (
	"aibuddy/internal/model"
	aiuserService "aibuddy/internal/services/aiuser"
	"aibuddy/pkg/ahttp"
	"aibuddy/pkg/baidu"
	"aibuddy/pkg/config"
	logger "aibuddy/pkg/log"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = func() trace.Tracer {
	return otel.Tracer(config.Instance.Tracer.ServiceName)
}

// Handler 用户相关处理器
type Handler struct {
	UserServer *aiuserService.Service
}

// New 创建用户处理器实例
func New() *Handler {
	return &Handler{
		UserServer: aiuserService.New(),
	}
}

// Login 手机号码登录
func (h *Handler) Login(state *ahttp.State, req *NewLoginRequest) error {
	ctx, span := tracer().Start(state.Ctx.Request().Context(), "phone_login")
	defer span.End()

	span.SetAttributes(attribute.String("wechat_code", req.WechatCode))
	span.SetAttributes(attribute.String("encrypted_data", req.EncryptedData))
	span.SetAttributes(attribute.String("iv", req.IV))
	span.SetAttributes(attribute.String("phone", req.Phone))
	span.SetAttributes(attribute.String("phone_code", req.PhoneCode))
	span.SetAttributes(attribute.String("source", req.Source))

	userInfo := &model.User{OpenID: "", Nickname: "", Phone: req.Phone, Avatar: ""}

	if req.Source == "phone" {
		// 验证码登陆
		if err := h.UserServer.CheckLoginCode(req.Phone, req.PhoneCode); err != nil {
			return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
		}
	} else {
		// 微信小程序登录
		wxUser, err := h.UserServer.CheckLoginMiniProgram(ctx, req.WechatCode, req.EncryptedData, req.IV, userInfo)
		if err != nil || wxUser == nil {
			return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
		}
		userInfo = wxUser
	}
	// 根据手机号获取用户信息
	user, err := h.UserServer.GetUserByPhone(state.Context(), userInfo.Phone)
	if err != nil {
		slog.Error(logger.Authorization, "msg", "Failed to get parent by phone", "error", err)
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}
	if user != nil {
		userInfo.ID = user.ID
	}
	if err := h.UserServer.UpsertUser(state.Context(), userInfo, req.Source); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error", err.Error()))
		span.SetAttributes(attribute.String("userinfo", userInfo.String()))
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	// 生成token并返回用户信息
	token, expires, err := aiuserService.GenerateToken(userInfo.ID, userInfo.Phone, userInfo.OpenID)
	if err != nil {
		slog.Error(logger.Authorization, "msg", "Failed to sign JWT token", "error", err)
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	return state.Resposne().SetData(LoginResponse{
		Token:    token,
		Expires:  expires,
		UID:      userInfo.ID,
		OpenID:   userInfo.OpenID,
		Nickname: userInfo.Nickname,
		Avatar:   userInfo.Avatar,
		Phone:    userInfo.Phone,
	}).Success()
}

// SendCode 验证码发送
func (h *Handler) SendCode(state *ahttp.State, req *SendCodeRequest) error {
	ctx, span := tracer().Start(state.Ctx.Request().Context(), "send_code")
	defer span.End()

	code, err := h.UserServer.SendPhoneCode(ctx, req.Phone)
	if err != nil {
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}
	slog.Info(logger.Authorization, "phone", req.Phone, "code", code)
	return state.Resposne().Success()
}

// RefreshToken 刷新token
func (h *Handler) RefreshToken(state *ahttp.State, req *TokenRequest) error {
	ctx, span := tracer().Start(state.Ctx.Request().Context(), "refresh_token")
	defer span.End()
	uid, err := aiuserService.GetUIDFromContext(state.Ctx)
	if err != nil {
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}
	token, expires, err := aiuserService.RefreshToken(ctx, req.Token, uid)
	if err != nil {
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	return state.Resposne().SetData(TokenResponse{
		Token:   token,
		Expires: expires,
	}).Success()
}

// Logout 用户退出登录
func (h *Handler) Logout(state *ahttp.State, _ *TokenRequest) error {
	return state.Resposne().Success()
}

// CompleteProfile 完善用户信息
func (h *Handler) CompleteProfile(state *ahttp.State, req *UserinfoRequest) error {
	ctx, span := tracer().Start(state.Ctx.Request().Context(), "complet_profile")
	defer span.End()

	uid, err := aiuserService.GetUIDFromContext(state.Ctx)
	if err != nil {
		span.RecordError(err)
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	birthday, err := time.Parse(time.DateOnly, req.Birthday)
	if err != nil {
		span.RecordError(err)
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	if err := h.UserServer.CompleteProfile(ctx, uid, req.BoardType, req.Relation, &model.DeviceInfo{
		ID:          req.ID,
		DeviceID:    req.DeviceID,
		NickName:    req.NickName,
		Avatar:      req.Avatar,
		Gender:      req.Gender,
		Birthday:    birthday,
		Hobbies:     []string{req.Hobbies},
		Values:      []string{req.Values},
		Skills:      []string{req.Skills},
		Personality: []string{req.Personality},
	}); err != nil {
		span.RecordError(err)
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	return state.Resposne().Success()
}

// Lost 发送挂失消息给设备
func (h *Handler) Lost(state *ahttp.State, req *LostRequest) error {
	ctx, span := tracer().Start(state.Context(), "Device.Lost")
	defer span.End()

	uid, err := aiuserService.GetUIDFromContext(state.Ctx)
	if err != nil {
		span.RecordError(err)
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	err = h.UserServer.Lost(ctx, uid, req.DeviceIDToUpper())
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", req.DeviceID))
		return state.Resposne().Error(err)
	}
	return state.Resposne().Success()
}

// Unlost 发送解除挂失消息给设备
func (h *Handler) Unlost(state *ahttp.State, req *UnlostRequest) error {
	ctx, span := tracer().Start(state.Context(), "Device.Unlost")
	defer span.End()

	err := h.UserServer.Unlost(ctx, req.DeviceIDToUpper())
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", req.DeviceID))
		return state.Resposne().Error(err)
	}
	return state.Resposne().Success()
}

// Unbind 发送解绑消息给设备
func (h *Handler) Unbind(state *ahttp.State, req *UnbindRequest) error {
	ctx, span := tracer().Start(state.Context(), "Device.Unbind")
	defer span.End()

	err := h.UserServer.Unbind(ctx, req.DeviceIDToUpper())
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", req.DeviceID))
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}
	return state.Resposne().Success()
}

// HaveDevice 用户是否已经绑定了设备
func (h *Handler) HaveDevice(state *ahttp.State) error {
	ctx, span := tracer().Start(state.Context(), "User.HaveDevice")
	defer span.End()

	uid, err := aiuserService.GetUIDFromContext(state.Ctx)
	if err != nil {
		span.RecordError(err)
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}
	haveDevice, err := h.UserServer.HaveDevice(ctx, uid)
	if err != nil {
		span.RecordError(err)
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	return state.Resposne().SetData(HaveDeviceResponse{
		HaveDevice: haveDevice,
	}).Success()
}

// DeviceList 设备列表
func (h *Handler) DeviceList(state *ahttp.State) error {
	ctx, span := tracer().Start(state.Context(), "User.DeviceList")
	defer span.End()

	uid, err := aiuserService.GetUIDFromContext(state.Ctx)
	if err != nil {
		span.RecordError(err)
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	deviceList, err := h.UserServer.UserDeviceList(ctx, uid)
	if err != nil {
		span.RecordError(err)
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	deviceListItems := make([]*DeviceInfoListItem, 0)
	for _, device := range deviceList {
		deviceName := ""
		avatar := ""
		gender := ""
		if device.DeviceInfo != nil {
			deviceName = device.DeviceInfo.NickName
			avatar = device.DeviceInfo.Avatar
			gender = device.DeviceInfo.Gender
		}
		deviceListItems = append(deviceListItems, &DeviceInfoListItem{
			DeviceID:   device.DeviceID,
			DeviceName: deviceName,
			Version:    device.Version,
			Status:     device.Status.String(),
			Avatar:     avatar,
			Gender:     gender,
		})
	}

	return state.Resposne().SetData(&DeviceListResponse{
		DeviceList: deviceListItems,
	}).Success()
}

// SendMessage 发送消息
func (h *Handler) SendMessage(state *ahttp.State, req *SendMsgRequest) error {
	ctx, span := tracer().Start(state.Context(), "User.SendMessage")
	defer span.End()

	uid, err := aiuserService.GetUIDFromContext(state.Ctx)
	if err != nil {
		span.RecordError(err)
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	if err := h.UserServer.CreateMessage(ctx, uid, &model.DeviceMessage{
		ToDeviceID: req.DeviceID,
		Fmt:        model.MessageFmt(req.Fmt),
		Content:    req.Content,
		Dur:        req.Dur,
	}); err != nil {
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	return state.Resposne().Success()
}

// MessageList 获取留言列表
func (h *Handler) MessageList(state *ahttp.State, req *GetMessageRequest) error {
	ctx, span := tracer().Start(state.Context(), "User.MessageList")
	defer span.End()

	uid, err := aiuserService.GetUIDFromContext(state.Ctx)
	if err != nil {
		span.RecordError(err)
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	data, total, err := h.UserServer.GetMessage(ctx, uid, req.DeviceID, req.Page, req.PageSize)
	if err != nil {
		span.RecordError(err)
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	return state.Resposne().SetData(MsgListResponse{
		Page:     req.Page,
		PageSize: req.PageSize,
		Total:    total,
		Data:     data,
	}).Success()
}

// AnalysisGrowthReport 分析用户成长报告
func (h *Handler) AnalysisGrowthReport(state *ahttp.State, req *AnalysisGrowthReportRequest) error {
	ctx, span := tracer().Start(state.Context(), "User.AnalysisGrowthReport")
	defer span.End()

	uid, err := aiuserService.GetUIDFromContext(state.Ctx)
	if err != nil {
		span.RecordError(err)
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	startTime, endTime, err := req.ParseTime()
	if err != nil {
		span.RecordError(err)
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	err = h.UserServer.AnalysisGrowthReport(ctx, uid, req.DeviceID, startTime, endTime)
	if err != nil {
		span.RecordError(err)
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	return state.Resposne().Success()
}

// GetGrowthReportList 获取用户成长报告列表
func (h *Handler) GetGrowthReportList(state *ahttp.State, req *GetGrowthReportListRequest) error {
	ctx, span := tracer().Start(state.Context(), "User.GetGrowthReportList")
	defer span.End()

	data, total, err := h.UserServer.GetGrowthReportList(ctx, req.DeviceID, req.Page, req.PageSize)
	if err != nil {
		span.RecordError(err)
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	return state.Resposne().SetData(GrowthReportListResponse{
		Page:     req.Page,
		PageSize: req.PageSize,
		Total:    total,
		Data:     data,
	}).Success()
}

// DeleteGrowthReport 删除成长报告
func (h *Handler) DeleteGrowthReport(state *ahttp.State, req *DeleteGrowthReportRequest) error {
	ctx, span := tracer().Start(state.Context(), "User.DeleteGrowthReport")
	defer span.End()

	err := h.UserServer.DeleteGrowthReport(ctx, req.ReportID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("report_id", req.ReportID))
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}
	return state.Resposne().Success()
}

// UpdateInfo 更新用户信息
func (h *Handler) UpdateInfo(state *ahttp.State, req *InfoRequest) error {
	ctx, span := tracer().Start(state.Context(), "User.GetGrowthReportList")
	defer span.End()
	uid, err := aiuserService.GetUIDFromContext(state.Ctx)
	if err != nil {
		span.RecordError(err)
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	user := &model.User{
		Nickname: req.NickName,
		Avatar:   req.Avatar,
		Phone:    req.Phone,
		Email:    req.Email,
		Username: req.Username,
		Gender:   *req.Gender,
	}
	if req.Birthday != "" {
		birthday, err := time.Parse(time.DateOnly, req.Birthday)
		if err != nil {
			span.SetAttributes(attribute.String("Birthday", req.Birthday))
			return state.Resposne().SetStatus(http.StatusBadRequest).Error(errors.New("生日日期参数格式错误"))
		}
		user.Birthday = birthday
	}

	if err := h.UserServer.UpdateUserInfo(ctx, uid, user); err != nil {
		span.RecordError(err)
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	return state.Resposne().Success()
}

// GetUserInfo 获取用户信息
func (h *Handler) GetUserInfo(state *ahttp.State) error {
	ctx, span := tracer().Start(state.Context(), "User.GetGrowthReportList")
	defer span.End()
	uid, err := aiuserService.GetUIDFromContext(state.Ctx)
	if err != nil {
		span.RecordError(err)
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	user, err := h.UserServer.GetUserInfoByUID(ctx, uid)
	if err != nil {
		span.RecordError(err)
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	birthday := ""
	if !user.Birthday.IsZero() {
		birthday = user.Birthday.Format(time.DateOnly)
	}

	return state.Resposne().SetData(InfoResponse{
		UID:      uid,
		Useranem: user.Username,
		Email:    user.Email,
		Phone:    user.Phone,
		Gender:   user.Gender,
		Birthday: birthday,
		NickName: user.Nickname,
		Avatar:   user.Avatar,
	}).Success()
}

// DeviceProfile 获取设备详细信息
func (h *Handler) DeviceProfile(state *ahttp.State, req *DeviceInfoRequest) error {
	ctx, span := tracer().Start(state.Context(), "User.DeviceProfile")
	defer span.End()
	uid, err := aiuserService.GetUIDFromContext(state.Ctx)
	if err != nil {
		span.RecordError(err)
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	info, device, err := h.UserServer.GetUserDeviceInfo(ctx, uid, req.DeviceID)
	if err != nil {
		span.RecordError(err)
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	return state.Resposne().SetData(&DeviceInfoResponse{
		Version:     device.Version,
		BoardType:   device.BoardType,
		DeviceID:    req.DeviceID,
		NickName:    info.NickName,
		Avatar:      info.Avatar,
		Gender:      info.Gender,
		Birthday:    info.Birthday.Format(time.DateOnly),
		Hobbies:     strings.Join(info.Hobbies, ","),
		Values:      strings.Join(info.Values, ","),
		Relation:    device.Relation,
		Skills:      strings.Join(info.Skills, ","),
		Personality: strings.Join(info.Personality, ","),
	}).Success()
}

// UpdateDeviceProfile 更新设备信息
func (h *Handler) UpdateDeviceProfile(state *ahttp.State, req *UpdateDeviceInfoRequest) error {
	ctx, span := tracer().Start(state.Context(), "User.UpdateDeviceProfile")
	defer span.End()

	uid, err := aiuserService.GetUIDFromContext(state.Ctx)
	if err != nil {
		span.RecordError(err)
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}
	birthday, err := time.Parse(time.DateOnly, req.Birthday)
	if err != nil {
		span.RecordError(err)
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}
	if err := h.UserServer.UpdateDeviceInfo(ctx, uid, &model.DeviceInfo{
		DeviceID:    req.DeviceID,
		NickName:    req.NickName,
		Avatar:      req.Avatar,
		Gender:      req.Gender,
		Birthday:    birthday,
		Hobbies:     []string{req.Hobbies},
		Values:      []string{req.Values},
		Skills:      []string{req.Skills},
		Personality: []string{req.Personality},
	}, req.Relation); err != nil {
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	return state.Resposne().Success()
}

// Unregister 注销用户帐号
func (h *Handler) Unregister(state *ahttp.State) error {
	ctx, span := tracer().Start(state.Context(), "User.Unregister")
	defer span.End()

	uid, err := aiuserService.GetUIDFromContext(state.Ctx)
	if err != nil {
		span.RecordError(err)
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	if err := h.UserServer.ClearUserInfo(ctx, uid); err != nil {
		span.RecordError(err)
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	return state.Resposne().Success()
}

// DownloadChatRecord 下载百度的聊天记录
func (h *Handler) DownloadChatRecord(state *ahttp.State, req *DownloadChatRecordRequest) error {
	_, span := tracer().Start(state.Context(), "User.DownloadChatRecord")
	defer span.End()

	if req.PageNo == 0 {
		req.PageNo = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	dialogues := baidu.NewDialogues()
	resp, err := dialogues.GetDialogues(&baidu.DialoguesRequest{
		UserID:    req.UserID,
		PageNo:    req.PageNo,
		PageSize:  req.PageSize,
		BeginTime: req.ToBeginTime(),
		EndTime:   req.ToEndTime(),
	})
	if err != nil {
		span.RecordError(err)
		return state.Resposne().Error(err)
	}

	return state.Resposne().SetData(DownloadChatRecordResponse{
		PageNo:   resp.PageNo,
		PageSize: resp.PageSize,
		Total:    len(resp.Data),
		Data:     resp.Data,
	}).Success()
}
