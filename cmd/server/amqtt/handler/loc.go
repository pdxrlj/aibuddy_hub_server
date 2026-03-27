// Package handler 基站定位处理器
package handler

import (
	"aibuddy/aiframe/location"
	"aibuddy/internal/query"
	"aibuddy/pkg/baidu"
	"aibuddy/pkg/mqtt"
	"fmt"
	"log/slog"
)

// LocHandler 基站定位处理器
type LocHandler struct {
}

// NewLocHandler 创建基站定位处理器
func NewLocHandler() *LocHandler {
	return &LocHandler{}
}

// Location 处理设备位置上报
func (h *LocHandler) Location(ctx *mqtt.Context) {
	defer ctx.Message.Ack()

	deviceID := ctx.Params["device_id"]
	slog.Info("[MQTT] Location", "device_id", deviceID, "payload", ctx.String())

	// 解析位置信息
	var loc location.Loc
	if err := loc.Decode(ctx.Payload); err != nil {
		slog.Error("[MQTT] Location decode failed", "device_id", deviceID, "error", err)
		return
	}

	// 验证必填参数
	if err := loc.Validate(); err != nil {
		slog.Error("[MQTT] Location validate failed", "device_id", deviceID, "error", err)
		return
	}

	// 执行定位
	resp, err := h.doLocate(&loc)
	if err != nil {
		slog.Error("[MQTT] Location API failed", "device_id", deviceID, "error", err)
		return
	}

	// 更新设备位置信息
	h.updateDeviceLocation(deviceID, resp)
}

// doLocate 执行定位请求
func (h *LocHandler) doLocate(loc *location.Loc) (*baidu.HardwareLocResponse, error) {
	// 从IMSI解析MCC和MNC
	imsiInfo, err := baidu.ParseIMSI(loc.IMSI)
	if err != nil {
		return nil, err
	}

	// 构建基站信息
	cell := baidu.BaseStation{
		MCC: imsiInfo.MCC,
		MNC: imsiInfo.MNC,
		LAC: loc.LAC,
		CID: loc.CID,
	}

	// 调用百度地图API进行定位
	client := baidu.NewMapsClient()
	switch loc.Source {
	case location.SourceTypeWifi:
		return client.SimpleWiFiLoc(loc.Data, loc.IMEI, cell, loc.CDMA, true)
	case location.SourceTypeBS:
		return client.SimpleCellLoc(loc.IMEI, cell, loc.CDMA, true)
	default:
		return nil, fmt.Errorf("unknown source type: %s", loc.Source)
	}
}

// updateDeviceLocation 更新设备位置信息到数据库
func (h *LocHandler) updateDeviceLocation(deviceID string, resp *baidu.HardwareLocResponse) {
	if resp == nil || !resp.IsSuccess() {
		errMsg := "unknown error"
		if resp != nil {
			errMsg = fmt.Sprintf("errcode=%d, msg=%s", resp.ErrCode, resp.ErrMsg)
		}
		slog.Error("[MQTT] Location API response error", "device_id", deviceID, "error", errMsg)
		return
	}

	longitude := fmt.Sprintf("%f", resp.Result.Lon)
	latitude := fmt.Sprintf("%f", resp.Result.Lat)
	locStr := resp.Result.GetAddress()

	slog.Info("[MQTT] Location success", "device_id", deviceID, "lat", latitude, "lon", longitude, "address", locStr)

	if _, err := query.Device.Where(query.Device.DeviceID.Eq(deviceID)).
		Updates(map[string]any{
			query.Device.Longitude.ColumnName().String(): longitude,
			query.Device.Latitude.ColumnName().String():  latitude,
			query.Device.Location.ColumnName().String():  locStr,
		}); err != nil {
		slog.Error("[MQTT] Location update failed", "device_id", deviceID, "error", err)
	}
}
