// Package ota 提供 OTA 固件升级相关功能
package ota

import (
	"aibuddy/aiframe"
	"aibuddy/pkg/mqtt"
	"encoding/json"
	"fmt"
)

// 	   "type": "ota_version",
//     "ota_url": "https://example.com/firmware/v2.0.0.bin",  // 升级包下载地址
//     "model_url":"xxxxx",
//     "ota_version": "2.0.0",        // 升级包版本号
//     "force_update": 1,              // 是否强制升级 (0-否, 1-是)
//     "force_update_model":1,// 模型是否更新
//     "update_content": "修复了xxx问题\n新增了xxx功能\n优化了xxx体验",  // 升级内容介绍
//     "device_id": "E4:B3:23:BF:C0:2C"

// Ota OTA 固件升级信息
type Ota struct {
	Type    string `json:"type"`
	Version string `json:"ver"`

	OtaURL      string `json:"ota_url"`
	ModelURL    string `json:"model_url"`
	ResourceURL string `json:"resource_url"`
	ForceUpdate bool   `json:"force"`
}

// Encode 将 OTA 信息序列化为 JSON
func (o *Ota) Encode() ([]byte, error) {
	return json.Marshal(o)
}

// Decode 从 JSON 数据反序列化到 OTA 结构
func (o *Ota) Decode(data []byte) error {
	return json.Unmarshal(data, o)
}

// String 将 OTA 信息转换为字符串
func (o *Ota) String() string {
	return fmt.Sprintf("Type: %s, Version: %s, OtaURL: %s, ModelURL: %s, ResourceURL: %s, ForceUpdate: %t", o.Type, o.Version, o.OtaURL, o.ModelURL, o.ResourceURL, o.ForceUpdate)
}

// SendToDevice 发送 OTA 信息到设备
func (o *Ota) SendToDevice(deviceID string) error {
	o.Type = "ota"
	topic := aiframe.OTATopic(deviceID)
	payload, err := o.Encode()
	if err != nil {
		return err
	}

	return mqtt.Instance.Publish(topic, 1, false, payload)
}
