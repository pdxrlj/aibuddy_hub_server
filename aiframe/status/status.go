// Package status 提供设备状态接口定义
package status

import (
	"encoding/json"
)

// DeviceStatus 设备状态
type DeviceStatus struct {
	Type    string `json:"type"`
	Battery int    `json:"bat"`      // 电池电量
	NetType string `json:"net_type"` // 网络类型 4g/wifi
}

// IsValid 验证设备状态是否有效
func (d *DeviceStatus) IsValid() bool {
	return d.Type == "stat"
}

// Encode 编码设备状态
func (d *DeviceStatus) Encode() ([]byte, error) {
	return json.Marshal(d)
}

// Decode 解码设备状态
func (d *DeviceStatus) Decode(data []byte) error {
	return json.Unmarshal(data, d)
}
