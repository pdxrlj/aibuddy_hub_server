// Package device 设备服务
package device

import (
	"aibuddy/aiframe/friend"
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
	"time"

	"github.com/spf13/cast"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = func() trace.Tracer {
	return otel.Tracer(config.Instance.Tracer.ServiceName)
}

// AfterConnectHook 连接后Hook
type AfterConnectHook func(ctx context.Context, deviceID string) error

// Service 设备服务
type Service struct {
	ClientIDPrefix         string
	cache                  flash.Flash
	DeviceRepo             *repository.DeviceRepo
	UserRepo               *repository.UserRepo
	DeviceRelationshipRepo *repository.DeviceRelationshipRepo

	DeviceMessageRepo *repository.DeviceMessageRepo
	OtaResource       *repository.DeviceOtaRepo
	OtaResourceRepo   *repository.OtaResourceRepo

	DeviceSnRepo *repository.BindDeviceSnRepo

	FileStorage storage.ObjectStorage[io.ReadCloser]

	// 连接后Hook
	AfterConnectHook []AfterConnectHook
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
		DeviceSnRepo:           repository.NewBindDeviceSnRepo(),
		OtaResource:            repository.NewDeviceOtaRepo(),
		OtaResourceRepo:        repository.NewOtaResourceRepo(),
		FileStorage: storage.NewStorage(
			config.Instance.Storage.OSS.AccessKeyID,
			config.Instance.Storage.OSS.AccessKeySecret,
			config.Instance.Storage.OSS.Region,
			config.Instance.Storage.OSS.Endpoint,
			config.Instance.Storage.OSS.Bucket,
		),
		AfterConnectHook: []AfterConnectHook{
			AfterConnectSendDeviceInfo,
		},
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
		slog.Info("[FirstOnline] GenerateAliyunMQTTAuth", "err", err.Error())
		return nil, err
	}

	span.SetAttributes(attribute.String("client_id", clientID))
	span.SetAttributes(attribute.String("username", username))
	span.SetAttributes(attribute.String("password", password))

	// 设备存在时更新版本号，不存在则忽略（新设备尚未绑定）
	if result, err := d.DeviceRepo.SetDeviceVersion(deviceID, version); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		slog.Error("[FirstOnline] SetDeviceVersion error", "err", err.Error())
		return nil, err
	} else if result.RowsAffected == 0 {
		slog.Info("[FirstOnline] device not found, skip version update", "device_id", deviceID)
	}

	// 为后续的完善用户信息做准备，缓存设备信息
	if err := d.cacheDeviceInfo(deviceID, simCard, version); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		slog.Error("[FirstOnline]", "cacheDeviceInfo error", err)
		return nil, err
	}

	mqttURL := mqttConfig.Mqtt.URL
	mqttURL = strings.Replace(mqttURL, "tcp", "mqtt", 1)

	for _, hook := range d.AfterConnectHook {
		if err := hook(ctx, deviceID); err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("device_id", deviceID))
			slog.Error("[FirstOnline]", "AfterConnectHook error", err)
			return nil, errors.New("无法发送用户信息给设备 " + err.Error())
		}
	}

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
	deviceID = strings.ToUpper(strings.ReplaceAll(deviceID, ":", "-"))
	return d.cache.Set("device_info:"+deviceID, jsonData)
}

// FromCacheGetDeviceInfo 获取设备 SIM 卡号和版本号信息
func (d *Service) FromCacheGetDeviceInfo(deviceID string) (simCard, version string, err error) {
	data, err := d.cache.Get("device_info:" + strings.ToUpper(strings.ReplaceAll(deviceID, ":", "-")))
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

// UseMQTTSendTargetDeviceToFriendInfo 查询设备信息后给对端发送好友信息,通过MQTT发送
func (d *Service) UseMQTTSendTargetDeviceToFriendInfo(ctx context.Context, deviceID, targetDeviceID string) error {
	slog.Info("[MQTT] UseMQTTSendTargetDeviceToFriendInfo", "deviceID", deviceID, "targetDeviceID", targetDeviceID)
	_, span := tracer().Start(ctx, "DeviceService.SendTargetDeviceToFriendInfo")
	defer span.End()

	deviceInfo, err := d.DeviceRepo.GetDeviceInfo(ctx, deviceID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID), attribute.String("target_device_id", targetDeviceID))
		return err
	}

	if deviceInfo.DeviceInfo == nil {
		return nil
	}

	isFriend, err := d.DeviceRelationshipRepo.IsFriend(ctx, deviceID, targetDeviceID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID), attribute.String("target_device_id", targetDeviceID))
		return err
	}

	relation := "朋友"
	if !isFriend {
		relation = "陌生人"
	}

	if err := friend.SendFriendIfo(deviceID, targetDeviceID, deviceInfo.DeviceInfo.NickName, deviceInfo.DeviceInfo.Avatar, relation); err != nil {
		slog.Warn("send friend info via mqtt failed", "error", err)
		return err
	}

	return nil
}

// AddFriend 添加好友
func (d *Service) AddFriend(ctx context.Context, deviceID, targetDeviceID string) (*model.Device, error) {
	_, span := tracer().Start(ctx, "DeviceService.AddFriend")
	defer span.End()
	targetDevice, err := d.DeviceRepo.GetDeviceInfo(ctx, targetDeviceID)
	if err != nil {
		slog.Error("[AddFriend]", "device_id", deviceID, "target_device_id", targetDeviceID, "error", err)
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID), attribute.String("target_device_id", targetDeviceID))
		return nil, err
	}

	if targetDevice.DeviceInfo == nil {
		slog.Error("[AddFriend]", "device_id", deviceID, "target_device_id", targetDeviceID, "error", "目标设备信息不存在")
		span.RecordError(errors.New("目标设备信息不存在"))
		span.SetAttributes(attribute.String("device_id", deviceID), attribute.String("target_device_id", targetDeviceID))
		return nil, errors.New("目标设备信息不存在")
	}
	// 判断是否已经是好友了
	isFriend, err := d.IsFriend(ctx, deviceID, targetDeviceID)
	if err != nil {
		slog.Error("[AddFriend]", "device_id", deviceID, "target_device_id", targetDeviceID, "error", err)
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID), attribute.String("target_device_id", targetDeviceID))
		return nil, err
	}

	if isFriend {
		return nil, errors.New("已经是好友关系，无法添加好友")
	}

	err = d.DeviceRelationshipRepo.CreateDeviceRelationship(ctx, deviceID, targetDeviceID, model.RelationshipStatusAccepted)
	if err != nil {
		slog.Error("[AddFriend]", "device_id", deviceID, "target_device_id", targetDeviceID, "error", err)
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID), attribute.String("target_device_id", targetDeviceID))
		return nil, err
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

// GetMessage 获取指定留言（按对话分组）
func (d *Service) GetMessage(ctx context.Context, deviceID string, page int, pageSize int) ([][]*MessageDTO, int64, error) {
	_, span := tracer().Start(ctx, "CreateMessage")
	defer span.End()

	messages, total, err := d.DeviceMessageRepo.GetMessageFromUser(ctx, deviceID, page, pageSize)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		return nil, 0, err
	}
	// helpers.PP(messages)

	dtoMessages := d.ToMessageDTO(messages)
	return dtoMessages, total, nil
}

// MessageDTO 留言响应DTO
type MessageDTO struct {
	ID           int    `json:"id"`
	MsgID        string `json:"msg_id"`
	FromDeviceID string `json:"from_device_id"`
	FromUsername string `json:"from_username"`
	ToDeviceID   string `json:"to_device_id"`
	FromAvatar   string `json:"from_avatar"`
	ToAvatar     string `json:"to_avatar"`
	ToUsername   string `json:"to_username"`
	Content      string `json:"content"`
	Fmt          string `json:"fmt"`
	Dur          int    `json:"dur"`
	Read         bool   `json:"read"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// ToMessageDTO 将 DeviceMessage 列表转换为按对话分组的 MessageDTO 列表
// 返回格式: [][]*MessageDTO，每个子数组代表与同一个聊天对象的所有消息
func (d *Service) ToMessageDTO(messages []*model.DeviceMessage) [][]*MessageDTO {
	// 按对话双方分组，确保 A->B 和 B->A 的消息在同一个组
	groups := make(map[string][]*MessageDTO)

	for _, msg := range messages {
		dto := &MessageDTO{
			ID:           msg.ID,
			MsgID:        msg.MsgID,
			FromDeviceID: msg.FromDeviceID,
			ToDeviceID:   msg.ToDeviceID,
			Content:      msg.Content,
			Fmt:          msg.Fmt.String(),
			Dur:          msg.Dur,
			Read:         msg.Read,
			CreatedAt:    time.Time(msg.CreatedAt).Format(time.DateTime),
			UpdatedAt:    time.Time(msg.UpdatedAt).Format(time.DateTime),
		}
		// 从 Device.DeviceInfo 获取头像
		if msg.Device != nil && msg.Device.DeviceInfo != nil {
			dto.FromAvatar = msg.Device.DeviceInfo.Avatar
			dto.FromUsername = msg.Device.DeviceInfo.NickName
		}
		// 从 ToDevice.DeviceInfo 获取头像
		if msg.ToDevice != nil && msg.ToDevice.DeviceInfo != nil {
			dto.ToAvatar = msg.ToDevice.DeviceInfo.Avatar
			dto.ToUsername = msg.ToDevice.DeviceInfo.NickName
		}

		// 生成分组key：将两个deviceID排序后拼接，确保双向对话在同一组
		key := makeConversationKey(msg.FromDeviceID, msg.ToDeviceID)
		groups[key] = append(groups[key], dto)
	}

	// 将 map 转换为二维数组
	result := make([][]*MessageDTO, 0, len(groups))
	for _, group := range groups {
		result = append(result, group)
	}

	return result
}

// makeConversationKey 生成分组key，确保 A-B 和 B-A 的对话使用相同的key
func makeConversationKey(id1, id2 string) string {
	if id1 < id2 {
		return id1 + ":" + id2
	}
	return id2 + ":" + id1
}

// AccountInfo 账户信息
type AccountInfo struct {
	NickName string `json:"nick_name"`
	Sex      string `json:"sex"`
	Birthday string `json:"birthday"`
	Sn       string `json:"sn"`
}

// GetAccountInfo 获取硬件的账户消息
func (d *Service) GetAccountInfo(ctx context.Context, deviceID string) (*AccountInfo, error) {
	_, span := tracer().Start(ctx, "DeviceService.GetAccountInfo")
	defer span.End()

	device, err := d.GetDeviceInfo(ctx, deviceID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		return nil, err
	}
	if device == nil || device.DeviceInfo == nil {
		span.RecordError(errors.New("无法查询到完整的设备信息"))
		span.SetAttributes(attribute.String("device_id", deviceID))
		return nil, errors.New("无法查询到完整的设备信息")
	}

	sn, err := d.DeviceSnRepo.GetDeviceSnByDeviceID(ctx, deviceID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		return nil, errors.New("无法查询到设备SN")
	}

	return &AccountInfo{
		NickName: device.DeviceInfo.NickName,
		Sex:      device.DeviceInfo.Gender,
		Birthday: device.DeviceInfo.Birthday.Format(time.DateOnly),
		Sn:       sn,
	}, nil
}

// OtaCheckResult OTA检查结果
type OtaCheckResult struct {
	NeedUpdate  bool   `json:"need_update"`  // 是否需要更新
	Version     string `json:"ver"`          // 最新版本号
	OtaURL      string `json:"ota_url"`      // OTA下载地址
	ModelURL    string `json:"model_url"`    // 模型下载地址
	ResourceURL string `json:"resource_url"` // 资源下载地址
	Force       bool   `json:"force"`        // 是否强制更新
}

// OtaCheck 升级校验
func (d *Service) OtaCheck(ctx context.Context, deviceID, currentVersion string) (*OtaCheckResult, error) {
	_, span := tracer().Start(ctx, "DeviceService.OtaCheck")
	defer span.End()

	_ = deviceID

	latestOta, err := d.OtaResourceRepo.GetLatestOtaResource(ctx, currentVersion)
	if err != nil {
		span.RecordError(err)
		slog.Error("[OtaCheck] get latest ota resource failed", "error", err)
		return nil, err
	}

	if latestOta == nil {
		return &OtaCheckResult{NeedUpdate: false}, nil
	}

	cmp := helpers.CompareVersion(currentVersion, latestOta.Version)
	if cmp >= 0 {
		return &OtaCheckResult{NeedUpdate: false}, nil
	}

	force := latestOta.ForceUpdate

	return &OtaCheckResult{
		NeedUpdate:  true,
		Version:     latestOta.Version,
		OtaURL:      latestOta.OtaURL,
		ModelURL:    latestOta.ModelURL,
		ResourceURL: latestOta.ResourceURL,
		Force:       force,
	}, nil
}

// SendMessageByName 根据接收消息的人的名称/或者父母名称发送消息
func (d *Service) SendMessageByName(ctx context.Context, deviceID, receiverName, content string) error {
	_, span := tracer().Start(ctx, "DeviceService.SendMessageByName")
	defer span.End()

	slog.Info("[SendMessageByName]", "device_id", deviceID, "receiver_name", receiverName)

	// 1. 获取发送设备信息
	deviceInfo, err := d.GetDeviceInfo(ctx, deviceID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		return errors.New("无法查询到设备信息")
	}
	if deviceInfo == nil || deviceInfo.DeviceInfo == nil {
		span.RecordError(errors.New("无法查询到完整的设备信息"))
		span.SetAttributes(attribute.String("device_id", deviceID))
		return errors.New("无法查询到完整的设备信息")
	}

	// 2. 先在好友列表中匹配 receiverName（好友昵称）
	if err = d.sendToFriend(ctx, deviceInfo, receiverName, content); err == nil {
		return nil
	}
	// 3. 好友中未匹配到，查询家庭成员（同一 UID 下的其他设备），匹配 relation 或昵称
	if deviceInfo.UID > 0 {
		if err = d.sendToFamily(ctx, deviceInfo, receiverName, content); err == nil {
			return nil
		}
	}

	span.RecordError(errors.New("未找到匹配的接收人"))
	span.SetAttributes(attribute.String("receiver_name", receiverName))
	return errors.New("未找到匹配的接收人")
}

// sendToFriend 尝试在好友列表中按昵称匹配并发送 MQTT 消息，匹配成功返回 nil
func (d *Service) sendToFriend(ctx context.Context, deviceInfo *model.Device, receiverName, content string) error {
	friend, err := d.DeviceRelationshipRepo.GetFriendByNickName(ctx, deviceInfo.DeviceID, receiverName)
	if err != nil {
		return err
	}
	if friend == nil || friend.TargetDevice == nil {
		return errors.New("好友中未找到匹配的接收人")
	}
	msgID := helpers.GenerateNumber(10)
	if err = d.DeviceMessageRepo.CreateDeviceMessage(ctx, &model.DeviceMessage{
		MsgID:        msgID,
		FromDeviceID: deviceInfo.DeviceID,
		ToDeviceID:   friend.TargetDeviceID,
		Content:      content,
		Fmt:          model.MessageFmtText,
	}); err != nil {
		return errors.New("创建消息失败")
	}

	username := deviceInfo.DeviceInfo.NickName
	if err = mqttmessage.SendMessage(deviceInfo.DeviceID, username, friend.TargetDeviceID, msgID, content, model.MessageFmtText.String(), 0); err != nil {
		return err
	}
	return nil
}

// sendToFamily 尝试在家庭成员中按 relation 或昵称匹配并发送 WebSocket 消息，匹配成功返回 nil
func (d *Service) sendToFamily(ctx context.Context, deviceInfo *model.Device, receiverName, content string) error {
	familyDevices, err := d.DeviceRepo.GetUserDeviceList(ctx, deviceInfo.UID)
	if err != nil {
		return err
	}

	for _, familyDevice := range familyDevices {
		if familyDevice.Relation == receiverName {
			msgID := helpers.GenerateNumber(10)
			if err = d.DeviceMessageRepo.CreateDeviceMessage(ctx, &model.DeviceMessage{
				MsgID:        msgID,
				FromDeviceID: deviceInfo.DeviceID,
				ToDeviceID:   familyDevice.DeviceID,
				Content:      content,
				Fmt:          model.MessageFmtText,
			}); err != nil {
				return errors.New("创建消息失败")
			}

			// 家庭成员使用小程序，只通过 WebSocket 通知，不需要 MQTT
			username := deviceInfo.DeviceInfo.NickName
			msg := map[string]any{
				"msg_id":    msgID,
				"from":      deviceInfo.DeviceID,
				"from_user": username,
				"content":   content,
				"fmt":       model.MessageFmtText,
				"dur":       0,
			}
			message, err := json.Marshal(msg)
			if err != nil {
				return err
			}

			websocket.SendMessage(cast.ToString(deviceInfo.UID), &websocket.DeviceToUserFrame{
				Type:     websocket.FrameTypeDeviceMsg,
				DeviceID: deviceInfo.DeviceID,
				Message:  message,
			})
			return nil
		}
	}
	return errors.New("家庭成员中未找到匹配的接收人")
}
