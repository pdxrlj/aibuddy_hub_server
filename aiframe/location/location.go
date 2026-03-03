// Package location 提供位置信息接口定义
package location

import (
	"aibuddy/aiframe"
	"aibuddy/pkg/mqtt"
	"encoding/json"
)

// {
//     "type": "bs",
//     "lon": "116.397428",
//     "lat": "39.90923",
//     "lac": 12345,
//     "cid": 67890
// }

// {
//     "type": "wifi",
//     "data": [
//         "54:6c:3e:77:89:ab,-45",
//         "a4:5d:3c:12:34:56,-67"
//     ]
// }

// SourceType 位置类型
type SourceType string

// SourceTypeBS 基站位置类型
const SourceTypeBS SourceType = "bs"

// SourceTypeWifi WiFi位置类型
const SourceTypeWifi SourceType = "wifi"

// IsValid 验证来源类型是否有效
func (l SourceType) IsValid() bool {
	return l == SourceTypeBS || l == SourceTypeWifi
}

// String 转换为来源类型字符串
func (l SourceType) String() string {
	return string(l)
}

// Loc 位置信息
type Loc struct {
	Type       string     `json:"type,omitempty"`
	Source     SourceType `json:"source,omitempty"`
	Data       []string   `json:"data,omitempty"`
	Latitude   float64    `json:"lat,omitempty"`
	Longitude  float64    `json:"lon,omitempty"`
	LocationID int        `json:"lac,omitempty"`
	ContentID  int        `json:"cid,omitempty"`
}

// Encode 编码位置信息
func (l *Loc) Encode() ([]byte, error) {
	return json.Marshal(l)
}

// Decode 解码位置信息
func (l *Loc) Decode(data []byte) error {
	return json.Unmarshal(data, l)
}

// SendToDevice 发送位置信息到设备
func (l *Loc) SendToDevice(deviceID string) error {
	payload, err := l.Encode()
	if err != nil {
		return err
	}
	return mqtt.Instance.Publish(aiframe.MQTTLocationTopic(deviceID), 1, false, payload)
}
