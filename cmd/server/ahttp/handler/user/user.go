// Package userhandler 提供用户相关的 HTTP 处理器
package userhandler

import (
	"aibuddy/internal/model"
	"aibuddy/internal/services/auth"
	"aibuddy/internal/services/cache"
	"aibuddy/pkg/ahttp"
	"aibuddy/pkg/config"
	logger "aibuddy/pkg/log"
	"aibuddy/pkg/wechatservice"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/redis/go-redis/v9"
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
	_, span := tracer().Start(state.Ctx.Request().Context(), "wechat_login")
	defer span.End()

	span.SetAttributes(attribute.String("wechat_code", req.WechatCode))
	span.SetAttributes(attribute.String("encrypted_data", req.EncryptedData))
	span.SetAttributes(attribute.String("iv", req.IV))

	parentInfo := &model.User{
		OpenID:   "",
		Nickname: "",
		Phone:    req.Phone,
		Avatar:   "",
	}

	if req.Source == "phone" {
		// 验证码登陆
		if err := checkLoginCode(req.Phone, req.PhoneCode); err != nil {
			return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
		}
	} else {
		// 微信小程序登录
		if err := checkLoginMiniProgram(req.WechatCode, req.EncryptedData, req.IV, parentInfo); err != nil {
			return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
		}
	}

	// 根据手机号获取用户信息
	parent, err := h.AuthServer.GetParentByPhone(state.Ctx.Request().Context(), parentInfo.Phone)
	if err != nil {
		slog.Error(logger.Authorization, "msg", "Failed to get parent by phone", "error", err)
		return state.Resposne().SetStatus(http.StatusInternalServerError).Error(err)
	}

	if parent == nil {
		// 如果用户不存在，创建新用户
		uid, err := h.AuthServer.CreateUser(state.Ctx.Request().Context(), parentInfo)
		if err != nil {
			slog.Error(logger.Authorization, "msg", "Failed to create user", "error", err)
			return state.Resposne().SetStatus(http.StatusInternalServerError).Error(err)
		}
		parentInfo.ID = uid
	} else {
		// 如果用户已存在，更新用户信息
		err = h.AuthServer.UpdateUser(state.Ctx.Request().Context(), parent.ID, parentInfo, parent)
		if err != nil {
			slog.Error(logger.Authorization, "msg", "Failed to update user", "error", err)
			return state.Resposne().SetStatus(http.StatusInternalServerError).Error(err)
		}
		parentInfo.ID = parent.ID
	}

	// 生成token并返回用户信息
	token, err := auth.GenerateToken(parentInfo.ID, parentInfo.Phone, parentInfo.OpenID)
	if err != nil {
		slog.Error(logger.Authorization, "msg", "Failed to sign JWT token", "error", err)
		return state.Resposne().SetStatus(http.StatusInternalServerError).Error(err)
	}

	return state.Resposne().SetData(LoginResponse{
		Token:    token,
		UID:      parentInfo.ID,
		OpenID:   parentInfo.OpenID,
		Nickname: parentInfo.Nickname,
		Phone:    parentInfo.Phone,
	}).Success()
}

// checkLoginCode 验证手机号登录验证码
func checkLoginCode(phone, code string) error {
	if phone == "" || code == "" {
		return errors.New("手机号码登录必要参数缺失")
	}

	// redis验证码验证
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

// checkLoginMiniProgram 验证微信小程序登录参数
func checkLoginMiniProgram(code, encryptedData, iv string, parentInfo *model.User) error {
	if code == "" || encryptedData == "" || iv == "" {
		return errors.New("小程序登录必要参数缺失")
	}
	// 获取微信小程序实例
	miniprogram, err := wechatservice.GetWechatMiniProgram()
	if err != nil {
		slog.Error(logger.Authorization, "msg", "Failed to get WeChat mini program instance", "error", err)
		return errors.New("failed to get WeChat mini program instance")
	}

	// 调用微信登录接口获取 session 信息
	session, err := miniprogram.GetAuth().Code2Session(code)
	if err != nil {
		slog.Error(logger.Authorization, "msg", "Failed to exchange WeChat code for session", "error", err)
		return errors.New("登录参数不合法")
	}

	// 获取用户手机号
	plainData, err := miniprogram.GetEncryptor().Decrypt(session.SessionKey, encryptedData, iv)
	if err != nil {
		slog.Error(logger.Authorization, "msg", "Failed to decrypt WeChat encrypted data", "error", err)
		return errors.New("参数异常")
	}

	parentInfo.OpenID = session.OpenID
	parentInfo.Phone = plainData.PhoneNumber
	parentInfo.Nickname = plainData.NickName
	parentInfo.Avatar = plainData.AvatarURL

	return nil
}
