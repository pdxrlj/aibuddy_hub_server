// Package handler 提供设备状态上报处理
package handler

import (
	"aibuddy/aiframe/status"
	"aibuddy/pkg/mqtt"
	"log/slog"
)

// StateHandler 设备状态上报处理
type StateHandler struct {
}

// NewStateHandler 创建设备状态上报处理
func NewStateHandler() *StateHandler {
	return &StateHandler{}
}

// Report 处理设备状态上报
func (h *StateHandler) Report(ctx *mqtt.Context) {
	defer ctx.Message.Ack()

	deviceID := ctx.Params["device_id"]
	var msg status.DeviceStatus

	if err := ctx.BindJSON(&msg); err != nil {
		ctx.Message.Ack()
		slog.Error("[MQTT] BindJSON failed", "error", err)
		return
	}

	_ = deviceID
	// TODO 设备状态保存到数据库
}
