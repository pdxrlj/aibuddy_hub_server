// Package device provides the device service.
package device

import (
	"aibuddy/aiframe/child"
	"aibuddy/internal/query"
	"context"
	"errors"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
)

// AfterConnectSendDeviceInfo 连接后发送设备信息
func AfterConnectSendDeviceInfo(ctx context.Context, deviceID string) error {
	_, span := tracer().Start(ctx, "DeviceService.AfterConnectSendDeviceInfo")
	defer span.End()

	ChildDeviceInfo, err := query.DeviceInfo.Where(query.DeviceInfo.DeviceID.Eq(deviceID)).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		slog.Error("[AfterConnectSendDeviceInfo] GetChildDeviceInfo", "error", err.Error())
		return err
	}

	if ChildDeviceInfo == nil {
		slog.Info("[AfterConnectSendDeviceInfo] SendChildInfoToDevice Skip Send Device Info")
		return nil
	}

	DeviceSN, err := query.DeviceSN.Where(query.DeviceSN.DeviceID.Eq(deviceID)).First()
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		slog.Error("[AfterConnectSendDeviceInfo] GetDeviceSN", "error", err.Error())
		return errors.New("获取设备SN失败")
	}
	sn := ""
	if DeviceSN != nil {
		sn = DeviceSN.SN
	}

	if ChildDeviceInfo != nil {
		slog.Info("[AfterConnectSendDeviceInfo] SendChildInfoToDevice")
		if err := child.SendChildInfoToDevice(ctx, deviceID, &child.Info{
			NickName: ChildDeviceInfo.NickName,
			Sn:       sn,
			Sex:      ChildDeviceInfo.Gender,
			Birthday: ChildDeviceInfo.Birthday.Format(time.DateOnly),
		}); err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("device_id", deviceID))
			slog.Error("[AfterConnectSendDeviceInfo] SendChildInfoToDevice", "error", err.Error())
			return err
		}
		slog.Info("[AfterConnectSendDeviceInfo] SendChildInfoToDevice success")
		return nil
	}
	slog.Info("[AfterConnectSendDeviceInfo] SendChildInfoToDevice not found")
	return nil
}
