// Package adebug 调试工具
package adebug

import (
	aiuserService "aibuddy/internal/services/aiuser"
	"aibuddy/pkg/ahttp"
	"net/http"
	"time"
)

// Handler 调试处理器
type Handler struct{}

// NewHandler 创建调试处理器实例
func NewHandler() *Handler {
	return &Handler{}
}

// ParseToken 解析Token
func (h *Handler) ParseToken(state *ahttp.State, req *ParseTokenRequest) error {
	claims, err := aiuserService.ValidateToken(state.Ctx, req.Token)
	if err != nil {
		return state.Response().SetStatus(http.StatusBadRequest).Error(err)
	}

	return state.Response().SetData(&ParseTokenResponse{
		UID:    claims.UID,
		Phone:  claims.Phone,
		OpenID: claims.OpenID,
		Exp:    time.Unix(claims.ExpiresAt, 0).Format(time.DateTime),
		Iat:    time.Unix(claims.IssuedAt, 0).Format(time.DateTime),
	}).Success()
}
