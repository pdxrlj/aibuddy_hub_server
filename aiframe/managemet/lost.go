// Package management 挂失与解绑管理
package management

import (
	"aibuddy/aiframe"
	"aibuddy/pkg/mqtt"
	"encoding/json"
)

// Lost 挂失请求
type Lost struct {
	Type    string `json:"type"`              // lost or unlost
	Contact string `json:"contact,omitempty"` // 联系人
	Phone   string `json:"phone,omitempty"`   // 联系手机号
}

// Encode 编码
func (l *Lost) Encode() ([]byte, error) {
	return json.Marshal(l)
}

// Decode 解码
func (l *Lost) Decode(data []byte) error {
	return json.Unmarshal(data, l)
}

// SendLost 发送挂失请求
func SendLost(deviceID, contact, phone string) error {
	topic := aiframe.MQTTLostTopic(deviceID)
	lost := &Lost{
		Type:    MgmtTypeLost.String(),
		Contact: contact,
		Phone:   phone,
	}

	payload, err := lost.Encode()

	if err != nil {
		return err
	}

	return mqtt.Instance.Publish(topic, 1, false, payload)
}

// SendUnlost 发送解除挂失请求
func SendUnlost(deviceID string) error {
	topic := aiframe.MQTTUnLostTopic(deviceID)
	unlost := &Lost{
		Type: MgmtTypeUnlost.String(),
	}

	payload, err := unlost.Encode()
	if err != nil {
		return err
	}

	return mqtt.Instance.Publish(topic, 1, false, payload)
}
