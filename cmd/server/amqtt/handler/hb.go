// Package handler 提供心跳上报处理
package handler

import (
	"aibuddy/aiframe/hb"
	"aibuddy/internal/model"
	"aibuddy/internal/repository"
	"aibuddy/internal/services/websocket"
	"aibuddy/pkg/mqtt"
	"context"
	"encoding/json"
	"log/slog"

	"github.com/spf13/cast"
)

// HbHandler 心跳上报处理器
type HbHandler struct {
	deviceRepo *repository.DeviceRepo
}

// NewHbHandler 创建心跳处理器
func NewHbHandler() *HbHandler {
	h := &HbHandler{
		deviceRepo: repository.NewDeviceRepo(),
	}

	return h
}

// Handle 处理心跳消息
func (h *HbHandler) Handle(ctx *mqtt.Context) {
	defer ctx.Message.Ack()

	deviceID := ctx.Params["device_id"]

	var msg hb.Hb
	if err := msg.Decode(ctx.Payload); err != nil {
		slog.Error("[MQTT] HbHandler decode failed", "device_id", deviceID, "error", err)
		return
	}

	if err := h.deviceRepo.SetDeviceStatus(deviceID, model.DeviceStatusOnline); err != nil {
		slog.Error("[MQTT] HbHandler set device status failed", "device_id", deviceID, "error", err)
		return
	}

	hardware := struct {
		Battery int    `json:"bat"`      // 电池电量
		NetType string `json:"net_type"` // 网络类型 4g/wifi
	}{
		Battery: msg.Battery,
		NetType: msg.NetType,
	}
	hardwareJSON, _ := json.Marshal(hardware)
	if err := h.deviceRepo.UpdateDeviceHardwareInfo(deviceID, hardwareJSON); err != nil {
		slog.Error("[MQTT] HbHandler update hardware info failed", "device_id", deviceID, "error", err)
	}

	slog.Info("[MQTT] HbHandler heartbeat received", "device_id", deviceID)

	device, err := h.deviceRepo.GetDeviceInfo(context.Background(), deviceID)
	if err != nil {
		slog.Error("[MQTT] HbHandler get device info failed", "device_id", deviceID, "error", err)
		return
	}

	onlineFrame := &websocket.DeviceOnlineFrame{
		DeviceID: device.DeviceID,
		Type:     websocket.FrameTypeOnline,
		Message:  hardwareJSON,
	}
	websocket.SendMessage(cast.ToString(device.UID), onlineFrame)
}
