// Package websocket 提供 websocket 处理服务
package websocket

import (
	"aibuddy/internal/services/aiuser"
	"aibuddy/internal/services/websocket"
	"aibuddy/pkg/ahttp"
	"errors"
	"log/slog"
)

// Handler Websocket处理器
type Handler struct {
	Service *websocket.Service
}

// NewHandler 创建Websocket处理器
func NewHandler() *Handler {
	return &Handler{Service: websocket.NewWebsocket()}
}

// HandleConnect 处理连接
func (h *Handler) HandleConnect(state *ahttp.State, req *HandleConnectRequest) error {
	// Token
	claims, err := aiuser.ValidateToken(state.Ctx, req.Token)
	if err != nil {
		slog.Error("[Websocket] HandleConnect", "error", err)
		return state.Resposne().Error(err)
	}

	if claims == nil {
		slog.Error("[Websocket] HandleConnect", "error", errors.New("token is invalid"))
		return state.Resposne().Error(errors.New("无法获取有效的用户信息"))
	}

	uid := claims.UID

	if err := h.Service.HandleConnect(uid, state.Ctx.Response().Writer, state.Ctx.Request()); err != nil {
		slog.Error("[Websocket] HandleConnect", "error", err)
		// melody 已经写入了响应，直接返回 nil 避免重复写入
		return nil
	}
	return nil
}
