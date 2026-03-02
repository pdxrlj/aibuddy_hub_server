// Package userhandler 提供用户相关的 HTTP 处理器
package userhandler

import (
	"aibuddy/internal/model"
	"aibuddy/internal/services/auth"
	"aibuddy/pkg/ahttp"
	"aibuddy/pkg/config"
	logger "aibuddy/pkg/log"
	"log/slog"
	"net/http"
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
	AuthServer *auth.Service
}

// New 创建用户处理器实例
func New() *Handler {
	return &Handler{
		AuthServer: auth.New(),
	}
}

// Login 手机号码登录
func (h *Handler) Login(state *ahttp.State, req *NewLoginRequest) error {
	_, span := tracer().Start(state.Ctx.Request().Context(), "phone_login")
	defer span.End()

	span.SetAttributes(attribute.String("wechat_code", req.WechatCode))
	span.SetAttributes(attribute.String("encrypted_data", req.EncryptedData))
	span.SetAttributes(attribute.String("iv", req.IV))

	userInfo := &model.User{
		OpenID:   "",
		Nickname: "",
		Phone:    req.Phone,
		Avatar:   "",
	}

	if req.Source == "phone" {
		// 验证码登陆
		if err := h.AuthServer.CheckLoginCode(req.Phone, req.PhoneCode); err != nil {
			return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
		}
	} else {
		// 微信小程序登录
		wxUser, err := h.AuthServer.CheckLoginMiniProgram(req.WechatCode, req.EncryptedData, req.IV, userInfo)
		if err != nil || wxUser == nil {
			return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
		}
		userInfo = wxUser
	}
	// 根据手机号获取用户信息
	user, err := h.AuthServer.GetUserByPhone(state.Context(), userInfo.Phone)
	if err != nil {
		slog.Error(logger.Authorization, "msg", "Failed to get parent by phone", "error", err)
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	if user != nil {
		userInfo.ID = user.ID
	}
	if err := h.AuthServer.UpsertUser(state.Context(), userInfo); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error", err.Error()))
		span.SetAttributes(attribute.String("userinfo", userInfo.String()))
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	// 生成token并返回用户信息
	token, expires, err := auth.GenerateToken(userInfo.ID, userInfo.Phone, userInfo.OpenID)
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
		Phone:    userInfo.Phone,
	}).Success()
}

// SendCode 验证码发送
func (h *Handler) SendCode(state *ahttp.State, req *SendCodeRequest) error {
	_, span := tracer().Start(state.Ctx.Request().Context(), "send_code")
	defer span.End()

	code, err := h.AuthServer.SendPhoneCode(req.Phone)
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

	token, expires, err := auth.RefreshToken(ctx, req.Token)
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

	uid, err := auth.GetUIDFromContext(state.Ctx)
	if err != nil {
		span.RecordError(err)
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	birthday, err := time.Parse(time.DateOnly, req.Birthday)
	if err != nil {
		span.RecordError(err)
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	if err := h.AuthServer.CompleteProfile(ctx, uid, &model.DeviceInfo{
		ID:          req.ID,
		DeviceID:    req.DeviceID,
		NickName:    req.NickName,
		Avatar:      req.Avatar,
		Gender:      req.Gender,
		Birthday:    birthday,
		Relation:    req.Relation,
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
