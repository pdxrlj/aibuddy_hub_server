// Package handler 基站定位处理器
package handler

import (
	"aibuddy/aiframe/location"
	"aibuddy/internal/query"
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
	defer ctx.Message.Ack()

	deviceID := ctx.Params["device_id"]
	slog.Info("[MQTT] Location", "device_id", deviceID, "payload", ctx.String())

	var loc location.Loc
	if err := loc.Decode(ctx.Payload); err != nil {
		slog.Error("[MQTT] Location", "device_id", deviceID, "error", err)
		return
	}

	// TODO: 保存位置信息到数据库

	longitude := loc.Longitude
	latitude := loc.Latitude
	location := ""

	if _, err := query.Device.Where(query.Device.DeviceID.Eq(deviceID)).
		Updates(map[string]any{
			query.Device.Longitude.ColumnName().String(): longitude,
			query.Device.Latitude.ColumnName().String():  latitude,
			query.Device.Location.ColumnName().String():  location,
		}); err != nil {
		slog.Error("[MQTT] Location", "device_id", deviceID, "error", err)
		return
	}
}
