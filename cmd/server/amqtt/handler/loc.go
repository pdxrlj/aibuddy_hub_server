// Package handler 基站定位处理器
package handler

import (
	"aibuddy/pkg/mqtt"
	"log/slog"
)

// LocHandler 基站定位处理器
type LocHandler struct {
}

// NewLocHandler 创建基站定位处理器
func NewLocHandler() *LocHandler {
	return &LocHandler{}
}

// Location 处理基站定位
func (h *LocHandler) Location(ctx *mqtt.Context) {
	deviceID := ctx.Params["device_id"]
	slog.Info("[MQTT] Location", "device_id", deviceID, "payload", ctx.String())
}
