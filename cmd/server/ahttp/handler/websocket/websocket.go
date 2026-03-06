// Package websocket 提供 websocket 处理服务
package websocket

import (
	"aibuddy/internal/services/aiuser"
	"aibuddy/internal/services/websocket"
	"aibuddy/pkg/ahttp"
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
func (h *Handler) HandleConnect(state *ahttp.State) error {
	uid, err := aiuser.GetUIDFromContext(state.Ctx)
	if err != nil {
		slog.Error("[Websocket] HandleConnect", "error", err)
		return state.Resposne().Error(err)
	}

	if err := h.Service.HandleConnect(uid, state.Ctx.Response(), state.Ctx.Request()); err != nil {
		slog.Error("[Websocket] HandleConnect", "error", err)
		return state.Resposne().Error(err)
	}
	return nil
}
