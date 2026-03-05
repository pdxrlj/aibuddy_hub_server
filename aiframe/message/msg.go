// Package message 实现了好友与留言相关功能
package message

import (
	"aibuddy/aiframe"
	"aibuddy/pkg/mqtt"
	"encoding/json"
)

// MsgType 消息类型
type MsgType string

const (
	// MsgTypeRecv 接收留言
	MsgTypeRecv MsgType = "recv"
)

// IsValid 验证消息类型是否有效
func (m MsgType) IsValid() bool {
	return m == MsgTypeRecv
}

// String 转换为字符串
func (m MsgType) String() string {
	return string(m)
}

// RecvMsg 接收留言消息
type RecvMsg struct {
	Type     MsgType `json:"type"`
	Mid      string  `json:"mid,omitempty"`
	From     string  `json:"from,omitempty"`
	FromName string  `json:"from_name,omitempty"`
	Fmt      string  `json:"fmt,omitempty"`
	Content  string  `json:"content,omitempty"`
	Dur      int     `json:"dur,omitempty"`
}

// SendMessage 发送消息
func SendMessage(fromDeviceID, fromUsername, targetDeviceID, msgID, content string, fmt string, dur int) error {
	msg := &RecvMsg{
		Type:     MsgTypeRecv,
		Mid:      msgID,
		From:     fromDeviceID,
		FromName: fromUsername,
		Fmt:      fmt,
		Content:  content,
		Dur:      dur,
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	topic := aiframe.MQTTSendMessageTopic(targetDeviceID)

	return mqtt.Instance.Publish(topic, 1, false, payload)
}
