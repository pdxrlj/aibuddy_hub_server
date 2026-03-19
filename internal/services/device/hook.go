// Package device provides the device service.
package device

import (
	"aibuddy/aiframe/child"
	"aibuddy/internal/query"
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

// AfterConnectSendDeviceInfo 连接后发送设备信息
func AfterConnectSendDeviceInfo(ctx context.Context, deviceID string) error {
	_, span := tracer().Start(ctx, "DeviceService.AfterConnectSendDeviceInfo")
	defer span.End()

	ChildDeviceInfo, err := query.DeviceInfo.Where(query.DeviceInfo.DeviceID.Eq(deviceID)).First()
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		return err
	}

	if ChildDeviceInfo != nil {
		return child.SendChildInfoToDevice(ctx, deviceID, &child.Info{
			NickName: ChildDeviceInfo.NickName,
			Sex:      ChildDeviceInfo.Gender,
			Birthday: ChildDeviceInfo.Birthday.Format(time.DateOnly),
		})
	}

	return nil
}
