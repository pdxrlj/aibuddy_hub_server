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
	Title   string `json:"title"`
	Content string `json:"content"`
	Remarks string `json:"remarks"`
}

// SendMessage 发送消息
func SendMessage(msgType, mid, fmt, title, content, targetDeviceID, remarks string) error {
	msg := &Msg{
		Type:    msgType,
		Mid:     mid,
		Fmt:     fmt,
		Title:   title,
		Content: content,
		Remarks: remarks,
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	topic := aiframe.MQTTSendMessageTopic(targetDeviceID)

	return mqtt.Instance.Publish(topic, 1, false, payload)
}
