// Package friend 定义好友相关的消息类型
package friend

import (
	"aibuddy/aiframe"
	"encoding/json"
)

const (
	// GetFriendInfoType 获取好友信息类型
	GetFriendInfoType = "get_friend_info"
)

// Friend 好友信息
type Friend struct {
	Type       string `json:"type"`
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
	Avatar     string `json:"avatar"`
	Relation   string `json:"relation"`
}

// Encode 编码好友信息
func (f *Friend) Encode() ([]byte, error) {
	return json.Marshal(f)
}

// Decode 解码好友信息
func (f *Friend) Decode(data []byte) error {
	return json.Unmarshal(data, f)
}

// SendFriendIfo 发送好友信息
func SendFriendIfo(deviceID, targetDeviceID, deviceName, avatar, relation string) error {
	topic := aiframe.MQTTAddFriendTopic(deviceID)

	message := &Friend{
		Type:       GetFriendInfoType,
		DeviceID:   targetDeviceID,
		DeviceName: deviceName,
		Avatar:     avatar,
		Relation:   relation,
	}

	data, err := message.Encode()
	if err != nil {
		return err
	}

	return aiframe.PublishToDevice(topic, data)
}
