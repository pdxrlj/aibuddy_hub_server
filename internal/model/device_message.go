// Package model defines the models for the device message.
package model

import (
	"aibuddy/pkg/config"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

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

	Device   *Device `gorm:"foreignKey:FromDeviceID;references:DeviceID" json:"device,omitempty"`
	ToDevice *Device `gorm:"foreignKey:ToDeviceID;references:DeviceID" json:"to_device,omitempty"`

	Content   string     `gorm:"column:content;" json:"content"`
	Fmt       MessageFmt `gorm:"column:fmt;not null" json:"fmt"`
	Dur       int        `gorm:"column:dur;" json:"dur"`
	Read      bool       `gorm:"column:read;not null;default:false" json:"read"`
	CreatedAt LocalTime  `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt LocalTime  `gorm:"column:updated_at;not null" json:"updated_at"`
}

// TableName 返回数据库表名
func (d *DeviceMessage) TableName() string {
	return TableName("device_message")
}

// AfterFind 在查询到设备消息后，将语音消息URL转换为完整的URL
func (d *DeviceMessage) AfterFind(_ *gorm.DB) error {
	if d.Fmt == MessageFmtVoice {
		domainname := DefaultDomainName
		if config.Instance != nil && config.Instance.App != nil && config.Instance.App.DomainName != "" {
			domainname = config.Instance.App.DomainName
		}

		d.Content = fmt.Sprintf("%s/api/v1/file/%s/file_proxy?filename=%s", domainname, d.FromDeviceID, d.Content)
	}
	return nil
}

// BeforeCreate 在插入之前
func (d *DeviceMessage) BeforeCreate(_ *gorm.DB) (err error) {
	if d.FromDeviceID != "" {
		d.FromDeviceID = strings.ToUpper(d.FromDeviceID)
	}

	if d.ToDeviceID != "" {
		d.ToDeviceID = strings.ToUpper(d.ToDeviceID)
	}

	return nil
}
