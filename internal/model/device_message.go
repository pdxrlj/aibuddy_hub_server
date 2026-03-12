// Package model defines the models for the device message.
package model

import "time"

// MessageFmt 消息格式
type MessageFmt string

// MessageFmtText 文本消息
const (
	// MessageFmtText 文本消息
	MessageFmtText MessageFmt = "text"
	// MessageFmtVoice 语音消息
	MessageFmtVoice MessageFmt = "voice"
)

// IsValid 是否有效
func (m MessageFmt) IsValid() bool {
	return m == MessageFmtText || m == MessageFmtVoice
}

// String 转换为字符串
func (m MessageFmt) String() string {
	return string(m)
}

// DeviceMessage 设备消息
type DeviceMessage struct {
	ID int `gorm:"column:id;primaryKey;autoIncrement" json:"id"`

	MsgID        string `gorm:"column:msg_id;not null" json:"msg_id"`
	FromDeviceID string `gorm:"column:from_device_id;not null" json:"from_device_id"`
	FromUsername string `gorm:"column:from_username;not null" json:"from_username"`

	ToDeviceID string `gorm:"column:to_device_id;not null" json:"to_device_id"`

	Device   *Device `gorm:"foreignKey:FromDeviceID;references:DeviceID" json:"device"`
	ToDevice *Device `gorm:"foreignKey:ToDeviceID;references:DeviceID" json:"to_device"`

	Content   string     `gorm:"column:content;" json:"content"`
	Fmt       MessageFmt `gorm:"column:fmt;not null" json:"fmt"`
	Dur       int        `gorm:"column:dur;" json:"dur"`
	Read      bool       `gorm:"column:read;not null;default:false" json:"read"`
	CreatedAt time.Time  `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at;not null" json:"updated_at"`
}

// TableName 返回数据库表名
func (d *DeviceMessage) TableName() string {
	return TableName("device_message")
}
