package handler

import (
	ai "aibuddy/aiframe/ai_chat"
	"aibuddy/internal/services/cache"
	"aibuddy/pkg/flash"
	"aibuddy/pkg/mqtt"
	"fmt"
	"log/slog"
	"time"
)

// AiChatHandler AI对话处理器
type AiChatHandler struct {
	cache flash.Flash
}

// NewAiChatHandler 创建AI对话处理器
func NewAiChatHandler() *AiChatHandler {
	return &AiChatHandler{
		cache: cache.Flash(),
	}
}

// Chat 处理AI对话
func (h *AiChatHandler) Chat(ctx *mqtt.Context) {
	defer ctx.Message.Ack()

	deviceID := ctx.Params["device_id"]
	var msg ai.Chat

	if err := msg.Decode(ctx.Payload); err != nil {
		slog.Error("[MQTT] Decode failed", "error", err)
		return
	}

	// 角色切换
	if msg.Type == ai.ChatTypeRole {
		slog.Error("[MQTT] Invalid chat type", "type", msg.Type)
		return
	}

	cacheKey := fmt.Sprintf("ai_chat:%s:%s:%s", deviceID, msg.Sid, msg.Type)
	if err := h.cache.Set(cacheKey, msg, time.Hour*24); err != nil {
		slog.Error("[MQTT] Set cache failed", "error", err)
		return
	}

	// TODO 处理 AI 对话
}
