// Package remind 提醒事件协议
package remind

import (
	"aibuddy/aiframe"
	"aibuddy/pkg/mqtt"
	"encoding/json"
)

const (
	// MsgTypeRemind 事件提醒
	MsgTypeRemind = "remind"
)

// Msg 通知相关信息
type Msg struct {
	Type    string `json:"type"`
	Mid     string `json:"mid"`
	Fmt     string `json:"fmt"`
	Sender  string `json:"sender"`
	From    string `json:"from"`
	Content string `json:"content"`
}

// SendMessage 发送消息
func SendMessage(mid, fmt, sender, content, from, targetDeviceID string) error {
	msg := &Msg{
		Type:    MsgTypeRemind,
		Mid:     mid,
		Fmt:     fmt,
		Sender:  sender,
		From:    from,
		Content: content,
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	topic := aiframe.MQTTSendMessageTopic(targetDeviceID)

	return mqtt.Instance.Publish(topic, 1, false, payload)
}
