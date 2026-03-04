// Package handler 提供心跳上报处理
package handler

import (
	"aibuddy/aiframe/hb"
	"aibuddy/internal/model"
	"aibuddy/internal/repository"
	"aibuddy/pkg/mqtt"
	"encoding/json"
	"log/slog"
)

// HbHandler 心跳上报处理器
type HbHandler struct {
	deviceRepo *repository.DeviceRepo
}

// NewHbHandler 创建心跳处理器
func NewHbHandler() *HbHandler {
	return &HbHandler{
		deviceRepo: repository.NewDeviceRepo(),
	}
}

// Handle 处理心跳消息
func (h *HbHandler) Handle(ctx *mqtt.Context) {
	defer ctx.Message.Ack()

	deviceID := ctx.Params["device_id"]

	var msg hb.Hb
	if err := msg.Decode(ctx.Payload); err != nil {
		slog.Error("[MQTT] HbHandler", "device_id", deviceID, "error", err)
		return
	}

	if err := h.deviceRepo.SetDeviceStatus(deviceID, model.DeviceStatusOnline); err != nil {
		slog.Error("[MQTT] HbHandler", "device_id", deviceID, "error", err)
		return
	}

	hardware := struct {
		Battery int    `json:"bat"`      // 电池电量
		NetType string `json:"net_type"` // 网络类型 4g/wifi
	}{
		Battery: msg.Battery,
		NetType: msg.NetType,
	}

	hardwareJSON, err := json.Marshal(hardware)
	if err != nil {
		slog.Error("[MQTT] HbHandler", "device_id", deviceID, "error", err)
		return
	}

	if err := h.deviceRepo.UpdateDeviceHardwareInfo(deviceID, hardwareJSON); err != nil {
		slog.Error("[MQTT] HbHandler", "device_id", deviceID, "error", err)
		return
	}
}
