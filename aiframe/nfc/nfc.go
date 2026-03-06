// Package nfc 定义NFC相关的消息类型
package nfc

import (
	"aibuddy/aiframe"
	"aibuddy/pkg/mqtt"
	"encoding/json"
)

// Type 定义NFC消息类型
type Type string

const (
	// NFCCreateType 卡制作请求
	NFCCreateType = "nfc_create"
	// NFCCreateResType 卡制完成响应
	NFCCreateResType = "nfc_created"
)

// CrateMessage NFC卡制作消息
type CrateMessage struct {
	Type  Type   `json:"type"`
	Cid   string `json:"cid"`
	Ctype string `json:"ctype"`
}

// Encode 编码NFC卡制作消息
func (m *CrateMessage) Encode() ([]byte, error) {
	return json.Marshal(m)
}

// Decode 解码NFC卡制作消息
func (m *CrateMessage) Decode(data []byte) error {
	return json.Unmarshal(data, m)
}

// CreateResRequest NFC卡制作响应请求
type CreateResRequest struct {
	Type  Type   `json:"type"`
	NFCID string `json:"nfc_id"`

	Cid   string `json:"cid"`
	Ctype string `json:"ctype"`
}

// Encode 编码NFC卡制作响应请求
func (m *CreateResRequest) Encode() ([]byte, error) {
	return json.Marshal(m)
}

// Decode 解码NFC卡制作响应请求
func (m *CreateResRequest) Decode(data []byte) error {
	return json.Unmarshal(data, m)
}

// SendNFCCreate 发送NFC卡制作消息
func SendNFCCreate(deviceID, cid, ctype string) error {
	topic := aiframe.NFCCrateTopic(deviceID)
	message := &CrateMessage{
		Type:  NFCCreateType,
		Cid:   cid,
		Ctype: ctype,
	}

	payload, err := message.Encode()
	if err != nil {
		return err
	}

	return mqtt.Instance.Publish(topic, 1, false, payload)
}
