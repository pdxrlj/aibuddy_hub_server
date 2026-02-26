// Package location 提供位置信息接口定义
package location

import "encoding/json"

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

// LocType 位置类型
type LocType string

// LocTypeBS 基站位置类型
const LocTypeBS LocType = "bs"

// LocTypeWifi WiFi位置类型
const LocTypeWifi LocType = "wifi"

// IsValid 验证位置类型是否有效
func (l LocType) IsValid() bool {
	return l == LocTypeBS || l == LocTypeWifi
}

// String 转换为字符串
func (l LocType) String() string {
	return string(l)
}

// Loc 位置信息
type Loc struct {
	Type       LocType  `json:"type,omitempty"`
	Data       []string `json:"data,omitempty"`
	Latitude   float64  `json:"lat,omitempty"`
	Longitude  float64  `json:"lon,omitempty"`
	LocationID int      `json:"lac,omitempty"`
	ContentID  int      `json:"cid,omitempty"`
}

// Encode 编码位置信息
func (l *Loc) Encode() ([]byte, error) {
	return json.Marshal(l)
}

// Decode 解码位置信息
func (l *Loc) Decode(data []byte) error {
	return json.Unmarshal(data, l)
}
