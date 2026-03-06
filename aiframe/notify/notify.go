// Package notify 通知协议
package notify

import (
	"aibuddy/aiframe"
	"aibuddy/pkg/mqtt"
	"encoding/json"
)

const (
	// MsgTypeNotify 通知
	MsgTypeNotify = "notify"
)

// Msg 通知相关信息
type Msg struct {
	Type    string `json:"type"`
	Mid     string `json:"mid"`
	Sub     string `json:"sub"`
	Sender  string `json:"sender"`
	Content string `json:"content"`
}

// SendMessage 发送消息
func SendMessage(mid, sub, content, targetDeviceID string) error {
	msg := &Msg{
		Type:    MsgTypeNotify,
		Mid:     mid,
		Sub:     sub,
		Content: content,
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	topic := aiframe.MQTTSendMessageTopic(targetDeviceID)

	return mqtt.Instance.Publish(topic, 1, false, payload)
}
