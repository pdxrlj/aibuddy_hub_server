// Package device 设备服务
package device

import (
	"aibuddy/internal/query"
	"aibuddy/pkg/config"
	"aibuddy/pkg/helpers"
	"aibuddy/pkg/mqtt"
	"context"

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
}

// NewService 创建设备服务实例
func NewService() *Service {
	return &Service{
		ClientIDPrefix: "GID_AIBuddy@@@",
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
func (d *Service) FirstOnline(ctx context.Context, deviceID string) (*ConfigInfo, error) {
	_, span := tracer().Start(ctx, "DeviceService.FirstOnline")
	defer span.End()

	mqttConfig := config.Instance.Aliyun
	clientID := d.ClientIDPrefix + deviceID
	username, password, err := mqtt.GenerateAliyunMQTTAuth(clientID, mqttConfig.AccessKeyID, mqttConfig.AccessKeySecret, mqttConfig.Mqtt.InstanceID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		return nil, err
	}

	span.SetAttributes(attribute.String("client_id", clientID))
	span.SetAttributes(attribute.String("username", username))
	span.SetAttributes(attribute.String("password", password))

	result, err := query.Agent.Where(query.Agent.ID.Eq(1)).Find()
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	helpers.PP(result)

	return &ConfigInfo{
		MQTTURL:      mqttConfig.Mqtt.URL,
		InstanceID:   mqttConfig.Mqtt.InstanceID,
		MQTTUsername: username,
		MQTTPassword: password,
	}, nil
}
