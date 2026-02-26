// Package status 提供设备状态接口定义
package status

import "encoding/json"

// DeviceStatus 设备状态
type DeviceStatus struct {
	Type          string `json:"type"`
	Battery       int    `json:"bat"`          // 电池电量
	Charging      bool   `json:"charging"`     // 是否充电
	SignalQuality int    `json:"csq"`          // 信号质量
	StorageFree   int64  `json:"storage_free"` // 存储剩余空间
	MemFree       int64  `json:"mem_free"`     // 内存剩余空间
	Iccid         string `json:"iccid"`        // SIM卡ICCID
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
