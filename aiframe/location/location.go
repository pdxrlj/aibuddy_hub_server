// Package location 提供位置信息接口定义
package location

import (
	"aibuddy/aiframe"
	"aibuddy/pkg/mqtt"
	"encoding/json"
	"fmt"
)

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
	Type   string     `json:"type,omitempty"`            // loc（必填）
	Source SourceType `json:"source,omitempty"`          // wifi 或 bs（必填）
	IMEI   string     `json:"imei,omitempty"`            // 设备IMEI，15位数字（必填）
	IMSI   string     `json:"imsi,omitempty"`            // 设备IMSI，15位数字（必填，从中解析MCC和MNC）
	CDMA   int        `json:"cdma,omitempty"`            // 是否CDMA网络：0-否，1-是（必填）
	LAC    int        `json:"lac,omitempty"`             // 位置区域码（必填）
	CID    int        `json:"cid,omitempty"`             // 基站ID（必填）
	Data   []string   `json:"data,omitempty"`            // WiFi信息列表（source为wifi时必填）
	// 定位结果
	Latitude  float64 `json:"latitude,omitempty"`  // 纬度
	Longitude float64 `json:"longitude,omitempty"` // 经度
	Radius    float64 `json:"radius,omitempty"`    // 精度半径(米)
	Address   string  `json:"address,omitempty"`   // 地址
	Province  string  `json:"province,omitempty"`  // 省
	City      string  `json:"city,omitempty"`      // 市
	District  string  `json:"district,omitempty"`  // 区
}

// Validate 验证位置信息必填参数
func (l *Loc) Validate() error {
	if l.Type != "loc" {
		return fmt.Errorf("type必须为loc，当前: %s", l.Type)
	}
	if !l.Source.IsValid() {
		return fmt.Errorf("source无效，必须为wifi或bs，当前: %s", l.Source)
	}
	if l.IMEI == "" {
		return fmt.Errorf("imei不能为空")
	}
	if l.IMSI == "" {
		return fmt.Errorf("imsi不能为空")
	}
	if l.LAC == 0 {
		return fmt.Errorf("lac不能为空")
	}
	if l.CID == 0 {
		return fmt.Errorf("cid不能为空")
	}
	if l.Source == SourceTypeWifi && len(l.Data) == 0 {
		return fmt.Errorf("source为wifi时data不能为空")
	}
	return nil
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
