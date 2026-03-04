// Package hb 心跳上报处理器
package hb

import (
	"encoding/json"
)

// Hb 心跳上报数据结构
type Hb struct {
	DeviceID string `json:"device_id"`
	Type     string `json:"type"`
	Battery  int    `json:"bat"`      // 电池电量
	NetType  string `json:"net_type"` // 网络类型 4g/wifi
}

// Encode 序列化心跳数据为JSON
func (h *Hb) Encode() ([]byte, error) {
	return json.Marshal(h)
}

// Decode 从JSON反序列化心跳数据
func (h *Hb) Decode(data []byte) error {
	return json.Unmarshal(data, h)
}
