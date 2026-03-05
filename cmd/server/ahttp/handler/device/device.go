// Package devicehandler provides a device handler.
package devicehandler

import (
	"aibuddy/internal/repository"
	"aibuddy/internal/services/device"
	"aibuddy/pkg/ahttp"
	"aibuddy/pkg/config"

	"github.com/cespare/xxhash/v2"
	"github.com/spf13/cast"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = func() trace.Tracer {
	return otel.Tracer(config.Instance.Tracer.ServiceName)
}

// Device is a device handler.
type Device struct {
	Service  *device.Service
	UserRepo *repository.UserRepo
}

// NewDevice creates a new device handler.
func NewDevice() *Device {
	return &Device{
		Service:  device.NewService(),
		UserRepo: repository.NewUserRepo(),
	}
}

// FirstOnline 设备第一次上线返回设备的配置信息，如 mqtt 的连接信息
func (d *Device) FirstOnline(state *ahttp.State, req *FirstOnlineRequest) error {
	ctx, span := tracer().Start(state.Context(), "Device.FirstOnline")
	defer span.End()

	configInfo, err := d.Service.FirstOnline(ctx, req.DeviceID, req.ICCID, req.Version)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", req.DeviceID))
		return state.Resposne().Error(err)
	}

	instanceID := xxhash.Sum64String(req.DeviceID)

	return state.Resposne().Success(&FirstOnlineResponse{
		MQTTConfig: &MQTTConfig{
			MQTTURL:      configInfo.MQTTURL,
			InstanceID:   configInfo.InstanceID,
			MQTTUsername: configInfo.MQTTUsername,
			MQTTPassword: configInfo.MQTTPassword,
		},
		DeviceInfo: &DeviceInfo{
			UserID:     req.DeviceID,
			InstanceID: instanceID,
		},
	})
}

// GetLocation 获取设备位置
func (d *Device) GetLocation(state *ahttp.State, req *GetLocationRequest) error {
	ctx, span := tracer().Start(state.Context(), "Device.GetLocation")
	defer span.End()

	err := d.Service.GetLocation(ctx, req.DeviceID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", req.DeviceID))
		return state.Resposne().Error(err)
	}
	return state.Resposne().Success()
}

// GetFriends 获取好友列表
func (d *Device) GetFriends(state *ahttp.State, req *GetFriendsRequest) error {
	ctx, span := tracer().Start(state.Context(), "Device.GetFriends")
	defer span.End()

	friends, total, err := d.Service.GetFriends(ctx, req.DeviceID, req.Page, req.Size)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", req.DeviceID))
		return state.Resposne().Error(err)
	}

	// 把妈妈信息也添加上
	user, err := d.Service.FindUserInfoByDeviceID(ctx, req.DeviceID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", req.DeviceID))
		return state.Resposne().Error(err)
	}

	deviceInfo, err := d.Service.GetDeviceInfo(ctx, req.DeviceID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", req.DeviceID))
		return state.Resposne().Error(err)
	}

	friendsResponse := make([]*GetFriendsResponseItem, len(friends)+1)
	friendsResponse[0] = &GetFriendsResponseItem{
		DeviceID:   cast.ToString(user.ID),
		DeviceName: user.Nickname,
		Avatar:     user.Avatar,
		Relation:   deviceInfo.Relation,
	}

	for i, friend := range friends {
		var deviceName, avatar string
		if friend.TargetDevice.DeviceInfo != nil {
			deviceName = friend.TargetDevice.DeviceInfo.NickName
			avatar = friend.TargetDevice.DeviceInfo.Avatar
		}

		friendsResponse[i+1] = &GetFriendsResponseItem{
			DeviceID:   friend.TargetDeviceID,
			DeviceName: deviceName,
			Avatar:     avatar,
			Relation:   "朋友",
		}
	}

	return state.Resposne().Success(&GetFriendsResponse{
		Total:   total,
		Page:    req.Page,
		Size:    req.Size,
		Friends: friendsResponse,
	})
}
