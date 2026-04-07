package handler

import (
	"aibuddy/aiframe/message"
	"aibuddy/internal/repository"
	"aibuddy/pkg/mqtt"
	"context"
	"log/slog"
)

// MsgHandler 消息处理器
type MsgHandler struct {
	DeviceMsg *repository.DeviceMessageRepo
}

// NewMsgHandler 实例化消息处理器
func NewMsgHandler() *MsgHandler {
	return &MsgHandler{
		DeviceMsg: repository.NewDeviceMessageRepo(),
	}
}

// Handle 处理消息Handle
func (h *MsgHandler) Handle(ctx *mqtt.Context) {
	defer ctx.Message.Ack()

	deviceID := ctx.Params["device_id"]
	var msg message.HandMsg
	if err := msg.Decode(ctx.Payload); err != nil {
		slog.Error("[MQTT] MsgHandler decode failed", "device_id", deviceID, "error", err)
		return
	}
	slog.Info("[MQTT] MsgHandler Handle", "device_id", deviceID, "payload", msg.Mids)
	if err := h.DeviceMsg.BatchMessageRead(context.Background(), msg.Mids); err != nil {
		slog.Error("[MQTT] HbHandler read device message failed", "device_id", deviceID, "error", err)
		return
	}
}
