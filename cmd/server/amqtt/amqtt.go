package amqtt

import (
	"aibuddy/pkg/config"
	"aibuddy/pkg/mqtt"
	"context"
	"fmt"
)

// StartAMQTTServer 启动MQTT服务器
func StartAMQTTServer(ctx context.Context) error {
	cfg := config.Instance
	if cfg.Aliyun == nil || cfg.Aliyun.Mqtt == nil {
		return fmt.Errorf("aliyun mqtt config is nil")
	}

	mqttCfg := cfg.Aliyun.Mqtt

	// 构建 MQTT 配置
	mc := &mqtt.Config{
		URL:            mqttCfg.URL,
		InstanceID:     mqttCfg.InstanceID,
		TopicPrefix:    mqttCfg.TopicPrefix,
		CleanSession:   mqttCfg.CleanSession,
		KeepAlive:      mqttCfg.KeepAlive,
		ClientIDPrefix: mqttCfg.ClientIDPrefix,
		Aliyun: &mqtt.AliyunConfig{
			AccessKeyID:     cfg.Aliyun.Ak,
			AccessKeySecret: cfg.Aliyun.Sk,
		},
	}

	// 可选的重连配置
	if mqttCfg.Reconnect != nil {
		mc.Reconnect = &mqtt.ReconnectConfig{
			Delay: mqttCfg.Reconnect.Delay,
		}
	}

	mqttInstance, err := mqtt.Connect(mc)
	if err != nil {
		return fmt.Errorf("mqtt connect failed: %w", err)
	}

	// 初始化路由
	SetupRoutes(mqttInstance)

	go func() {
		<-ctx.Done()
		StopAMQTTServer()
	}()

	return nil
}

// StopAMQTTServer 停止MQTT服务器
func StopAMQTTServer() {
	if mqtt.Instance != nil {
		mqtt.Instance.Disconnect()
	}
}
