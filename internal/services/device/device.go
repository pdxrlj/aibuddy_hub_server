// Package device 设备服务
package device

import (
	"aibuddy/aiframe/location"
	"aibuddy/internal/services/cache"
	"aibuddy/pkg/config"
	"aibuddy/pkg/flash"
	"aibuddy/pkg/mqtt"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	ClientIDPrefix string
	cache          flash.Flash
}

// NewService 创建设备服务实例
func NewService() *Service {
	return &Service{
		ClientIDPrefix: "GID_AIBuddy@@@",
		cache:          cache.Flash(),
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
func (d *Service) FirstOnline(ctx context.Context, deviceID, iccid, version string) (*ConfigInfo, error) {
	_, span := tracer().Start(ctx, "DeviceService.FirstOnline")
	defer span.End()

	mqttConfig := config.Instance.Aliyun
	clientID := d.ClientIDPrefix + deviceID
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
	if err := d.cacheDeviceInfo(deviceID, iccid, version); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		return nil, err
	}

	return &ConfigInfo{
		MQTTURL:      mqttConfig.Mqtt.URL,
		InstanceID:   mqttConfig.Mqtt.InstanceID,
		MQTTUsername: username,
		MQTTPassword: password,
	}, nil
}

func (d *Service) cacheDeviceInfo(deviceID, iccid, version string) error {
	cacheData := map[string]string{
		"iccid":   iccid,
		"version": version,
	}
	jsonData, err := json.Marshal(cacheData)
	if err != nil {
		return err
	}
	deviceID = strings.ReplaceAll(deviceID, ":", "-")
	return d.cache.Set("device_info:"+deviceID, jsonData)
}

// FromCacheGetDeviceInfo 获取设备 ICCID 和版本号信息
func (d *Service) FromCacheGetDeviceInfo(deviceID string) (iccid, version string, err error) {
	data, err := d.cache.Get("device_info:" + strings.ReplaceAll(deviceID, ":", "-"))
	if err != nil {
		return "", "", fmt.Errorf("无法从缓存信息获取设备的 SN: %w", err)
	}

	var jsonData []byte
	switch v := data.(type) {
	case []byte:
		jsonData = v
	case string:
		jsonData = []byte(v)
	default:
		return "", "", fmt.Errorf("无法从缓存信息获取设备的 SN: 数据类型错误")
	}

	var cacheData map[string]string
	if err := json.Unmarshal(jsonData, &cacheData); err != nil {
		return "", "", errors.New("无法从缓存信息获取设备的 SN: 数据类型错误")
	}
	var ok bool

	iccid, ok = cacheData["iccid"]
	if !ok {
		return "", "", errors.New("无法从缓存信息获取设备的SN: iccid not found")
	}
	version, ok = cacheData["version"]
	if !ok {
		return "", "", errors.New("无法从缓存信息获取设备的SN: version not found")
	}

	return iccid, version, nil
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
