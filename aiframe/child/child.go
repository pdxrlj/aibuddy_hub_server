// Package child provides the child service.
package child

import (
	"aibuddy/aiframe"
	"context"
	"encoding/json"
)

// Info 子设备信息
type Info struct {
	Type     string `json:"type"`
	NickName string `json:"nick_name"`
	Sex      string `json:"sex"`
	Birthday string `json:"birthday"`
}

// Encode 编码
func (c *Info) Encode() ([]byte, error) {
	return json.Marshal(c)
}

// Decode 解码
func (c *Info) Decode(data []byte) error {
	return json.Unmarshal(data, c)
}

// SendChildInfoToDevice 发送子设备信息到设备
func SendChildInfoToDevice(_ context.Context, deviceID string, info *Info) error {
	if info.Type == "" {
		info.Type = "device_user"
	}

	payload, err := info.Encode()
	if err != nil {
		return err
	}
	return aiframe.PublishToDevice(aiframe.MQTTSetChildInfoTopic(deviceID), payload)
}
