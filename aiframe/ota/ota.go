// Package ota 提供 OTA 固件升级相关功能
package ota

import "encoding/json"

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
	Type             string `json:"type"`
	OtaURL           string `json:"ota_url"`
	ModelURL         string `json:"model_url"`
	OtaVersion       string `json:"ota_version"`
	ForceUpdate      int    `json:"force_update"`
	ForceUpdateModel int    `json:"force_update_model"`
	UpdateContent    string `json:"update_content"`
	DeviceID         string `json:"device_id"`
}

func (o *Ota) Encode() ([]byte, error) {
	return json.Marshal(o)
}

func (o *Ota) Decode(data []byte) error {
	return json.Unmarshal(data, o)
}
