// Package aiframe 提供 AI 设备框架接口定义
package aiframe

import (
	"aibuddy/pkg/config"
	"aibuddy/pkg/mqtt"
	"fmt"
)

// Frame 帧编码/解码接口
type Frame interface {
	Encode() ([]byte, error)
	Decode(data []byte) error
}

var (
	// MQTTBoundTopic 绑定主题
	MQTTBoundTopic = func(deviceID string) string {
		return GetTopic(fmt.Sprintf("%s/cmd/mgmt", deviceID))
	}

	// MQTTLocationTopic 位置主题
	MQTTLocationTopic = func(deviceID string) string {
		return GetTopic(fmt.Sprintf("%s/loc", deviceID))
	}

	// OTATopic OTA 主题
	OTATopic = func(deviceID string) string {
		return GetTopic(fmt.Sprintf("%s/ota", deviceID))
	}

	// MQTTUnbindTopic 发送解除绑定请求
	MQTTUnbindTopic = func(deviceID string) string {
		return GetTopic(fmt.Sprintf("%s/mgmt", deviceID))
	}

	// MQTTLostTopic 发送挂失请求
	MQTTLostTopic = func(deviceID string) string {
		return GetTopic(fmt.Sprintf("%s/mgmt", deviceID))
	}

	// MQTTUnLostTopic 发送解除挂失请求
	MQTTUnLostTopic = func(deviceID string) string {
		return GetTopic(fmt.Sprintf("%s/mgmt", deviceID))
	}
)

// GetTopic 获取 MQTT 主题
func GetTopic(topic string) string {
	topicPrefix := ""
	if config.Instance.Aliyun != nil && config.Instance.Aliyun.Mqtt != nil {
		topicPrefix = config.Instance.Aliyun.Mqtt.TopicPrefix
	}

	return mqtt.GetTopic(topicPrefix, topic)
}
