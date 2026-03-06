// Package amqtt 提供应用层MQTT路由配置
package amqtt

import (
	"aibuddy/cmd/server/amqtt/handler"
	"aibuddy/pkg/config"
	"aibuddy/pkg/mqtt"
)

// MsgNotification 消息通知结构体-只是测试一下
type MsgNotification struct {
	Type     string `json:"type"`
	DeviceID string `json:"device_id"`
	MsgType  int    `json:"msg_type"`
	SubType  int    `json:"sub_type"`
	Sender   string `json:"sender"`
	Content  string `json:"content"`
}

// SetupRoutes 设置路由
func SetupRoutes(instance *mqtt.Mqtt) {
	topicPrefix := ""
	if config.Instance.Aliyun != nil && config.Instance.Aliyun.Mqtt != nil {
		topicPrefix = config.Instance.Aliyun.Mqtt.TopicPrefix
	}

	mqtt.InitRouter(instance, topicPrefix, mqtt.WithDebug(true))

	r := mqtt.GetRouter()

	locHandler := handler.NewLocHandler()
	// 基站定位
	r.On(":device_id/loc", locHandler.Location)

	hbHandler := handler.NewHbHandler()
	// 心跳上报
	r.On(":device_id/hb", hbHandler.Handle)

	// AI 对话
	aiChatHandler := handler.NewAiChatHandler()
	r.On(":device_id/ai", aiChatHandler.Chat)

	nfcHandler := handler.NewNFCHandler()
	r.On(":device_id/nfc", nfcHandler.Handle)
	// 消息处理
	msgHandler := handler.NewMsgHandler()
	r.On(":device_id/msg", msgHandler.Handle)
}
