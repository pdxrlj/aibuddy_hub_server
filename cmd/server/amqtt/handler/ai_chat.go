package handler

import (
	ai "aibuddy/aiframe/ai_chat"
	"aibuddy/pkg/mqtt"
	"log/slog"
)

// AiChatHandler AI对话处理器
type AiChatHandler struct{}

// NewAiChatHandler 创建AI对话处理器
func NewAiChatHandler() *AiChatHandler {
	return &AiChatHandler{}
}

// Chat 处理AI对话
func (h *AiChatHandler) Chat(ctx *mqtt.Context) {
	defer ctx.Message.Ack()

	deviceID := ctx.Params["device_id"]
	var msg ai.Chat
	if err := ctx.BindJSON(&msg); err != nil {
		ctx.Message.Ack()
		slog.Error("[MQTT] BindJSON failed", "error", err)
		return
	}

	_ = deviceID
	// TODO 处理 AI 对话
}
