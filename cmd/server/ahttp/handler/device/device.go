// Package devicehandler provides a device handler.
package devicehandler

import (
	"aibuddy/internal/repository"
	"aibuddy/internal/services/device"
	"aibuddy/pkg/ahttp"
	"aibuddy/pkg/config"
	"log/slog"
	"strconv"

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
	Service       *device.Service
	UserRepo      *repository.UserRepo
	DeviceMsgRepo *repository.DeviceMessageRepo
}

// NewDevice creates a new device handler.
func NewDevice() *Device {
	return &Device{
		Service:       device.NewService(),
		UserRepo:      repository.NewUserRepo(),
		DeviceMsgRepo: repository.NewDeviceMessageRepo(),
	}
}

// FirstOnline 设备第一次上线返回设备的配置信息，如 mqtt 的连接信息
func (d *Device) FirstOnline(state *ahttp.State, req *FirstOnlineRequest) error {
	ctx, span := tracer().Start(state.Context(), "Device.FirstOnline")
	defer span.End()

	configInfo, err := d.Service.FirstOnline(ctx, req.DeviceID, req.SIMCard, req.Version)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", req.DeviceID))
		return state.Response().Error(err)
	}

	instanceID := xxhash.Sum64String(req.DeviceID)

	return state.Response().Success(&FirstOnlineResponse{
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
		return state.Response().Error(err)
	}
	return state.Response().Success()
}

// GetFriends 获取好友列表
func (d *Device) GetFriends(state *ahttp.State, req *GetFriendsRequest) error {
	ctx, span := tracer().Start(state.Context(), "Device.GetFriends")
	defer span.End()

	friends, total, err := d.Service.GetFriends(ctx, req.DeviceID, req.Page, req.Size)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", req.DeviceID))
		slog.Error("[GetFriends] GetFriends", "device_id", req.DeviceID, "error", err)
		return state.Response().Error(err)
	}

	// 把微信用户信息也添加上
	user, err := d.Service.FindUserInfoByDeviceID(ctx, req.DeviceID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", req.DeviceID))
		slog.Info("[GetFriends] FindUserInfoByDeviceID", "device_id", req.DeviceID, "error", err)
		return state.Response().Error(err)
	}

	deviceInfo, err := d.Service.GetDeviceInfo(ctx, req.DeviceID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", req.DeviceID))
		slog.Info("[GetFriends] GetDeviceInfo", "device_id", req.DeviceID, "error", err)
		return state.Response().Error(err)
	}

	friendsResponse := make([]*GetFriendsResponseItem, len(friends)+1)
	friendsResponse[0] = &GetFriendsResponseItem{
		DeviceID:   cast.ToString(user.ID),
		DeviceName: user.Nickname,
		Avatar:     user.Avatar,
		Relation:   deviceInfo.Relation,
	}

	for i, friend := range friends {
		var deviceName, avatar, sex string
		if friend.TargetDevice.DeviceInfo != nil {
			deviceName = friend.TargetDevice.DeviceInfo.NickName
			avatar = friend.TargetDevice.DeviceInfo.Avatar
			sex = friend.TargetDevice.DeviceInfo.Gender
		}

		friendsResponse[i+1] = &GetFriendsResponseItem{
			DeviceID:   friend.TargetDeviceID,
			DeviceName: deviceName,
			Sex:        sex,
			Avatar:     avatar,
			Relation:   "朋友",
		}
	}

	return state.Response().Success(&GetFriendsResponse{
		Total:   total,
		Page:    req.Page,
		Size:    req.Size,
		Friends: friendsResponse,
	})
}

// GetDeviceInfo 获取设备信息
func (d *Device) GetDeviceInfo(state *ahttp.State, req *GetDeviceInfoRequest) error {
	ctx, span := tracer().Start(state.Context(), "Device.GetDeviceInfo")
	defer span.End()

	deviceInfo, err := d.Service.GetDeviceInfo(ctx, req.TargetDeviceID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", req.DeviceID), attribute.String("target_device_id", req.TargetDeviceID))
		return state.Response().Error(err)
	}

	isFriend, err := d.Service.IsFriend(ctx, req.DeviceID, req.TargetDeviceID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", req.DeviceID), attribute.String("target_device_id", req.TargetDeviceID))
		return state.Response().Error(err)
	}

	relation := "朋友"
	if !isFriend {
		relation = "陌生人"
	}

	var nickName, avatar string
	if deviceInfo.DeviceInfo != nil {
		nickName = deviceInfo.DeviceInfo.NickName
		avatar = deviceInfo.DeviceInfo.Avatar
	}

	// 对端设备信息，使用MQTT协议发送给对端
	if err := d.Service.UseMQTTSendTargetDeviceToFriendInfo(ctx, req.DeviceID, req.TargetDeviceID); err != nil {
		slog.Warn("send target device friend info failed", "error", err)
	}

	return state.Response().Success(&GetDeviceInfoResponse{
		DeviceID:   deviceInfo.DeviceID,
		DeviceName: nickName,
		Avatar:     avatar,
		Relation:   relation,
	})
}

// AddFriend 添加好友
func (d *Device) AddFriend(state *ahttp.State, req *AddFriendRequest) error {
	ctx, span := tracer().Start(state.Context(), "Device.AddFriend")
	defer span.End()

	targetDevice, err := d.Service.AddFriend(ctx, req.DeviceID, req.TargetDeviceID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", req.DeviceID), attribute.String("target_device_id", req.TargetDeviceID))
		return state.Response().Error(err)
	}

	// 对端设备信息，使用MQTT协议发送给对端
	if err := d.Service.UseMQTTSendTargetDeviceToFriendInfo(ctx, req.TargetDeviceID, req.DeviceID); err != nil {
		slog.Warn("send target device friend info failed", "error", err)
	}

	return state.Response().Success(&AddFriendResponse{
		Name:     targetDevice.DeviceInfo.NickName,
		Avatar:   targetDevice.DeviceInfo.Avatar,
		DeviceID: targetDevice.DeviceID,
	})
}

// DeleteFriend 删除好友
func (d *Device) DeleteFriend(state *ahttp.State, req *DeleteFriendRequest) error {
	ctx, span := tracer().Start(state.Context(), "Device.DeleteFriend")
	defer span.End()

	err := d.Service.DeleteFriend(ctx, req.DeviceID, req.TargetDeviceID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", req.DeviceID), attribute.String("target_device_id", req.TargetDeviceID))
		return state.Response().Error(err)
	}

	return state.Response().Success()
}

// SendMessage 发送消息
func (d *Device) SendMessage(state *ahttp.State, req *SendMessageRequest) error {
	ctx, span := tracer().Start(state.Context(), "Device.SendMessage")
	defer span.End()

	// 判断发送对象是否为用户uid
	uid, err := strconv.Atoi(req.TargetDeviceID)
	if err != nil {
		slog.Info("[sendmessage]", "check device id", req.TargetDeviceID, "result", err.Error())
	}
	if uid > 0 {
		if err := d.Service.SendMessageToUser(ctx, req.DeviceID, uid, req.Content, req.Fmt, req.Dur); err != nil {
			return state.Response().Error(err)
		}
		return state.Response().Success()
	}

	err = d.Service.SendMessage(ctx, req.DeviceID, req.TargetDeviceID, req.Content, req.Fmt, req.Dur)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", req.DeviceID), attribute.String("target_device_id", req.TargetDeviceID))
		return state.Response().Error(err)
	}

	return state.Response().Success()
}

// MessageList 设备消息列表
func (d *Device) MessageList(state *ahttp.State, req *MessageListRequest) error {
	ctx, span := tracer().Start(state.Context(), "Device.MessageList")
	defer span.End()

	data, total, err := d.Service.GetMessage(ctx, req.DeviceID, req.Page, req.Size)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(
			attribute.String("device_id", req.DeviceID),
			attribute.Int("page", req.Page),
			attribute.Int("size", req.Size),
		)
		return state.Response().Error(err)
	}

	return state.Response().Success(&MessageListResponse{
		Page:  req.Page,
		Size:  req.Size,
		Total: total,
		List:  data,
	})
}

// AccountInfo 获取硬件的账户消息
func (d *Device) AccountInfo(state *ahttp.State, req *AccountInfoRequest) error {
	ctx, span := tracer().Start(state.Context(), "Device.AccountInfo")
	defer span.End()

	accountInfo, err := d.Service.GetAccountInfo(ctx, req.DeviceID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", req.DeviceID))
		return state.Response().Error(err)
	}

	return state.Response().Success(accountInfo)
}

// OtaCheck ota 升级校验
func (d *Device) OtaCheck(state *ahttp.State, req *OtaCheckRequest) error {
	ctx, span := tracer().Start(state.Context(), "Device.OtaCheck")
	defer span.End()

	device, err := d.Service.OtaCheck(ctx, req.DeviceID, req.Version)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", req.DeviceID))
		return state.Response().Error(err)
	}

	return state.Response().Success(device)
}

// SendMessageByName 根据用户或者好友/家庭名称发送消息---- 设备到设备端
func (d *Device) SendMessageByName(state *ahttp.State, req *SendMessageByNameRequest) error {
	ctx, span := tracer().Start(state.Context(), "Device.SendMessageByName")
	defer span.End()

	err := d.Service.SendMessageByName(ctx, req.DeviceID, req.ReceiverName, req.Content)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", req.DeviceID), attribute.String("receiver_name", req.ReceiverName))
		return state.Response().Error(err)
	}

	return state.Response().Success()
}
