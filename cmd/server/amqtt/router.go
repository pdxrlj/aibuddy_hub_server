// Package amqtt 提供应用层MQTT路由配置
package amqtt

import (
	"log/slog"

	"aibuddy/cmd/server/amqtt/handler"
	"aibuddy/pkg/config"
	"aibuddy/pkg/helpers"
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

	// ================================ 待验证 ================================

	nfcHandler := handler.NewNFCHandler()
	r.On(":device_id/nfc", nfcHandler.Handle)

	// 设备状态
	r.On("device/:id/status", func(ctx *mqtt.Context) {
		deviceID := ctx.Params["id"]
		msgID := ctx.Message.MessageID()
		var msg MsgNotification
		if err := ctx.BindJSON(&msg); err != nil {
			ctx.Message.Ack()
			slog.Error("[MQTT] BindJSON failed", "error", err)
			return
		}
		helpers.PP(msg)
		slog.Info("[MQTT] Device status========", "device_id", deviceID, "msg_id", msgID)
		ctx.Message.Ack()
		slog.Info("[MQTT] Message acknowledged==========", "device_id", deviceID, "msg_id", msgID)
	})

	// 设备请求
	r.On("device/:id/request", func(ctx *mqtt.Context) {
		deviceID := ctx.Params["id"]
		slog.Info("[MQTT] Device request", "device_id", deviceID, "payload", ctx.String())
		if err := ctx.Reply("device/"+deviceID+"/response", "ack"); err != nil {
			slog.Error("[MQTT] Reply failed", "error", err)
		}
	})
}
