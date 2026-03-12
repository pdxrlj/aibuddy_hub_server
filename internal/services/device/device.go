// Package device 设备服务
package device

import (
	"aibuddy/aiframe/location"
	mqttmessage "aibuddy/aiframe/message"
	"aibuddy/internal/model"
	"aibuddy/internal/repository"
	"aibuddy/internal/services/cache"
	"aibuddy/internal/services/websocket"
	"aibuddy/pkg/config"
	"aibuddy/pkg/flash"
	"aibuddy/pkg/helpers"
	"aibuddy/pkg/mqtt"
	"aibuddy/pkg/storage"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = func() trace.Tracer {
	return otel.Tracer(config.Instance.Tracer.ServiceName)
}

// Service 设备服务
type Service struct {
	ClientIDPrefix         string
	cache                  flash.Flash
	DeviceRepo             *repository.DeviceRepo
	UserRepo               *repository.UserRepo
	DeviceRelationshipRepo *repository.DeviceRelationshipRepo

	DeviceMessageRepo *repository.DeviceMessageRepo

	FileStorage storage.ObjectStorage[io.ReadCloser]
}

// NewService 创建设备服务实例
func NewService() *Service {
	if config.Instance.Storage == nil || config.Instance.Storage.OSS == nil {
		panic("storage config is not set")
	}

	return &Service{
		ClientIDPrefix:         "GID_AIBuddy@@@",
		cache:                  cache.Flash(),
		DeviceRepo:             repository.NewDeviceRepo(),
		DeviceRelationshipRepo: repository.NewDeviceRelationshipRepo(),
		UserRepo:               repository.NewUserRepo(),
		DeviceMessageRepo:      repository.NewDeviceMessageRepo(),
		FileStorage: storage.NewStorage(
			config.Instance.Storage.OSS.AccessKeyID,
			config.Instance.Storage.OSS.AccessKeySecret,
			config.Instance.Storage.OSS.Region,
			config.Instance.Storage.OSS.Endpoint,
			config.Instance.Storage.OSS.Bucket,
		),
	}
}

// ConfigInfo 设备配置信息
type ConfigInfo struct {
	MQTTURL      string `json:"mqtt_url"`
	InstanceID   string `json:"instance_id"`
	MQTTUsername string `json:"mqtt_username"`
	MQTTPassword string `json:"mqtt_password"`
}

// FirstOnline 设备第一次上线返回设备的配置信息，如 mqtt 的连接信息
// deviceID: 设备ID
// 返回：设备的配置信息，如 mqtt 的连接信息
// 错误：如果生成 MQTT 认证信息失败，则返回错误
func (d *Service) FirstOnline(ctx context.Context, deviceID, simCard, version string) (*ConfigInfo, error) {
	_, span := tracer().Start(ctx, "DeviceService.FirstOnline")
	defer span.End()

	mqttConfig := config.Instance.Aliyun
	clientID := d.ClientIDPrefix + deviceID
	slog.Info("[FirstOnline]", "client_id", clientID, "sim_card", simCard, "version", version)
	username, password, err := mqtt.GenerateAliyunMQTTAuth(clientID, mqttConfig.Ak, mqttConfig.Sk, mqttConfig.Mqtt.InstanceID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		return nil, err
	}

	span.SetAttributes(attribute.String("client_id", clientID))
	span.SetAttributes(attribute.String("username", username))
	span.SetAttributes(attribute.String("password", password))

	// 为后续的完善用户信息做准备，缓存设备信息
	if err := d.cacheDeviceInfo(deviceID, simCard, version); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		return nil, err
	}

	mqttURL := mqttConfig.Mqtt.URL
	mqttURL = strings.Replace(mqttURL, "tcp", "mqtt", 1)

	return &ConfigInfo{
		MQTTURL:      mqttURL,
		InstanceID:   mqttConfig.Mqtt.InstanceID,
		MQTTUsername: username,
		MQTTPassword: password,
	}, nil
}

func (d *Service) cacheDeviceInfo(deviceID, simCard, version string) error {
	cacheData := map[string]string{
		"sim_card": simCard,
		"version":  version,
	}
	jsonData, err := json.Marshal(cacheData)
	if err != nil {
		return err
	}
	deviceID = strings.ReplaceAll(deviceID, ":", "-")
	return d.cache.Set("device_info:"+deviceID, jsonData)
}

// FromCacheGetDeviceInfo 获取设备 SIM 卡号和版本号信息
func (d *Service) FromCacheGetDeviceInfo(deviceID string) (simCard, version string, err error) {
	data, err := d.cache.Get("device_info:" + strings.ReplaceAll(deviceID, ":", "-"))
	if err != nil {
		return "", "", fmt.Errorf("无法从缓存信息获取设备的 SIM 卡号: %w", err)
	}

	var jsonData []byte
	switch v := data.(type) {
	case []byte:
		jsonData = v
	case string:
		jsonData = []byte(v)
	default:
		return "", "", fmt.Errorf("无法从缓存信息获取设备的 SIM 卡号: 数据类型错误")
	}

	var cacheData map[string]string
	if err := json.Unmarshal(jsonData, &cacheData); err != nil {
		return "", "", errors.New("无法从缓存信息获取设备的 SIM 卡号: 数据类型错误")
	}
	var ok bool

	simCard, ok = cacheData["sim_card"]
	if !ok {
		return "", "", errors.New("无法从缓存信息获取设备的SIM卡号: sim_card not found")
	}
	version, ok = cacheData["version"]
	if !ok {
		return "", "", errors.New("无法从缓存信息获取设备的SIM卡号: version not found")
	}

	return simCard, version, nil
}

// GetLocation 获取设备位置信息
func (d *Service) GetLocation(ctx context.Context, deviceID string) error {
	_, span := tracer().Start(ctx, "DeviceService.GetLocation")
	defer span.End()

	loc := location.Loc{
		Type: "loc",
	}

	err := loc.SendToDevice(deviceID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		return err
	}
	return nil
}

// GetFriends 获取好友列表
func (d *Service) GetFriends(ctx context.Context, deviceID string, page, size int) ([]*model.DeviceRelationship, int64, error) {
	_, span := tracer().Start(ctx, "DeviceService.GetFriends")
	defer span.End()

	friends, total, err := d.DeviceRelationshipRepo.GetFriends(ctx, deviceID, page, size)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		return nil, 0, err
	}
	return friends, total, nil
}

// IsFriend 判断是否是好友关系
func (d *Service) IsFriend(ctx context.Context, deviceID, targetDeviceID string) (bool, error) {
	ctx, span := tracer().Start(ctx, "DeviceService.IsFirstOnline")
	defer span.End()

	isFriend, err := d.DeviceRelationshipRepo.IsFriend(ctx, deviceID, targetDeviceID)
	if err != nil {
		return false, err
	}

	return isFriend, nil
}

// FindUserInfoByDeviceID 根据设备ID查询用户信息
func (d *Service) FindUserInfoByDeviceID(ctx context.Context, deviceID string) (*model.User, error) {
	ctx, span := tracer().Start(ctx, "DeviceService.FindUserInfoByDeviceID")
	defer span.End()

	user, err := d.DeviceRepo.FindUserInfoByDeviceID(ctx, deviceID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		return nil, err
	}

	return user, nil
}

// GetDeviceInfo 获取设备信息
func (d *Service) GetDeviceInfo(ctx context.Context, deviceID string) (*model.Device, error) {
	_, span := tracer().Start(ctx, "DeviceService.GetDeviceInfo")
	defer span.End()

	device, err := d.DeviceRepo.GetDeviceInfo(ctx, deviceID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		return nil, err
	}

	return device, nil
}

// AddFriend 添加好友
func (d *Service) AddFriend(ctx context.Context, deviceID, targetDeviceID string) (*model.Device, error) {
	_, span := tracer().Start(ctx, "DeviceService.AddFriend")
	defer span.End()

	// 判断是否已经是好友了
	isFriend, err := d.IsFriend(ctx, deviceID, targetDeviceID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID), attribute.String("target_device_id", targetDeviceID))
		return nil, err
	}
	if isFriend {
		return nil, errors.New("已经是好友关系，无法添加好友")
	}

	err = d.DeviceRelationshipRepo.CreateDeviceRelationship(ctx, deviceID, targetDeviceID, model.RelationshipStatusAccepted)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID), attribute.String("target_device_id", targetDeviceID))
		return nil, err
	}

	targetDevice, err := d.DeviceRepo.GetDeviceInfo(ctx, targetDeviceID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID), attribute.String("target_device_id", targetDeviceID))
		return nil, err
	}

	if targetDevice.DeviceInfo == nil {
		span.RecordError(errors.New("目标设备信息不存在"))
		span.SetAttributes(attribute.String("device_id", deviceID), attribute.String("target_device_id", targetDeviceID))
		return nil, errors.New("目标设备信息不存在")
	}

	return targetDevice, nil
}

// DeleteFriend 删除好友
func (d *Service) DeleteFriend(ctx context.Context, deviceID, targetDeviceID string) error {
	_, span := tracer().Start(ctx, "DeviceService.DeleteFriend")
	defer span.End()

	err := d.DeviceRelationshipRepo.DeleteDeviceRelationship(ctx, deviceID, targetDeviceID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID), attribute.String("target_device_id", targetDeviceID))
		return err
	}
	return nil
}

// SendMessage 发送消息
func (d *Service) SendMessage(ctx context.Context, deviceID, targetDeviceID, content string, fmt string, dur int) error {
	_, span := tracer().Start(ctx, "DeviceService.SendMessage")
	defer span.End()

	slog.Info("[SendMessage]", "device_id", deviceID, "target_device_id", targetDeviceID)
	// 确认是好友关系
	isFriend, err := d.IsFriend(ctx, deviceID, targetDeviceID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID), attribute.String("target_device_id", targetDeviceID))
		return errors.New("确认好友关系失败")
	}
	if !isFriend {
		span.RecordError(errors.New("不是好友关系"))
		span.SetAttributes(attribute.String("device_id", deviceID), attribute.String("target_device_id", targetDeviceID))
		return errors.New("不是好友关系，无法发送消息")
	}
	msgID := helpers.GenerateNumber(10)
	deviceInfo, err := d.GetDeviceInfo(ctx, deviceID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		return err
	}

	if deviceInfo == nil || deviceInfo.DeviceInfo == nil {
		span.RecordError(errors.New("无法查询到完整的设备信息"))
		span.SetAttributes(attribute.String("device_id", deviceID))
		return errors.New("无法查询到完整的设备信息")
	}

	if err = d.DeviceMessageRepo.CreateDeviceMessage(ctx, &model.DeviceMessage{
		MsgID:        msgID,
		FromDeviceID: deviceID,
		FromUsername: deviceInfo.DeviceInfo.NickName,
		ToDeviceID:   targetDeviceID,
		Content:      content,
		Fmt:          model.MessageFmt(fmt),
		Dur:          dur,
	}); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID), attribute.String("target_device_id", targetDeviceID))
		return errors.New("创建消息失败")
	}

	username := deviceInfo.DeviceInfo.NickName
	err = mqttmessage.SendMessage(deviceID, username, targetDeviceID, msgID, content, fmt, dur)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID), attribute.String("target_device_id", targetDeviceID))
		return err
	}

	return nil
}

// SendMessageToUser 发送信息通过ID
func (d *Service) SendMessageToUser(ctx context.Context, deviceID string, uid int, content string, fmt string, dur int) error {
	ctx, span := tracer().Start(ctx, "DeviceService.SendMessageToUser")
	defer span.End()
	info, err := d.DeviceRepo.GetDeviceInfo(ctx, deviceID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID), attribute.Int("uid", uid))
		return err
	}

	if int(info.UID) != uid {
		span.SetAttributes(attribute.Int64("device_id", info.UID), attribute.Int("uid", uid))
		return errors.New("无发送给用户信息的权限")
	}

	msgID := helpers.GenerateNumber(10)
	deviceInfo, err := d.GetDeviceInfo(ctx, deviceID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		return err
	}

	if deviceInfo == nil || deviceInfo.DeviceInfo == nil {
		span.RecordError(errors.New("无法查询到完整的设备信息"))
		span.SetAttributes(attribute.String("device_id", deviceID))
		return errors.New("无法查询到完整的设备信息")
	}

	if err = d.DeviceMessageRepo.CreateDeviceMessage(ctx, &model.DeviceMessage{
		MsgID:        msgID,
		FromDeviceID: deviceID,
		FromUsername: deviceInfo.DeviceInfo.NickName,
		ToDeviceID:   strconv.Itoa(uid),
		Content:      content,
		Fmt:          model.MessageFmt(fmt),
		Dur:          dur,
	}); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID), attribute.Int("target_", uid))
		return errors.New("创建消息失败")
	}

	msg := map[string]any{
		"msg_id":    msgID,
		"from":      deviceID,
		"from_user": deviceInfo.DeviceInfo.NickName,
		"content":   content,
		"fmt":       model.MessageFmt(fmt),
		"dur":       dur,
	}
	message, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	websocket.SendMessage(strconv.Itoa(uid), &websocket.DeviceToUserFrame{
		Type:     websocket.FrameTypeDeviceMsg,
		DeviceID: deviceID,
		Message:  message,
	})

	return nil
}
