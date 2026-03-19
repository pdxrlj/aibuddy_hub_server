// Package aiuser provides the user service.
package aiuser

import (
	"aibuddy/aiframe/child"
	"aibuddy/internal/query"
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

// AfterCompleteProfileHook 完善资料后Hook
type AfterCompleteProfileHook func(ctx context.Context, deviceID string) error

// AfterCompleteProfileSendChildInfo 完善资料后发送子设备信息
func AfterCompleteProfileSendChildInfo(ctx context.Context, deviceID string) error {
	_, span := tracer().Start(ctx, "AiuserService.AfterCompleteProfileSendChildInfo")
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
