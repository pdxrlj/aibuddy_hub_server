// Package model provides the model for the NFC card
package model

import (
	"aibuddy/pkg/config"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"gorm.io/gorm"
)

// NFCStatus NFC状态
type NFCStatus string

const (
	// NFCPending 制作中
	NFCPending NFCStatus = "制作中"

	// NFCPaid 制作完成
	NFCPaid NFCStatus = "制作完成"

	// NFCInvalid 已失效
	NFCInvalid NFCStatus = "已失效"
)

// String 返回NFC状态字符串
func (s NFCStatus) String() string {
	return string(s)
}

// NFC 模型
type NFC struct {
	ID       int64  `gorm:"column:id;type:bigint;autoIncrement;primaryKey" json:"id"`
	DeviceID string `gorm:"column:device_id;type:varchar(255);not null" json:"device_id"`
	UID      int64  `gorm:"column:uid;type:bigint;" json:"-"`
	Cid      string `gorm:"column:cid;type:varchar(255);not null" json:"cid"`
	Ctype    string `gorm:"column:ctype;type:varchar(255);not null" json:"ctype"`
	NFCID    string `gorm:"column:nfc_id;type:varchar(255);not null" json:"nfc_id"`

	Title   string    `gorm:"column:title;type:varchar(255);not null" json:"title"`
	Content string    `gorm:"column:content;type:text;not null" json:"content"`
	Voice   string    `gorm:"column:voice;type:varchar(255)" json:"voice"`
	Picture string    `gorm:"column:picture;type:varchar(255);" json:"picture"`
	Status  NFCStatus `gorm:"column:status;type:varchar(255);not null;default:'制作中'" json:"status"`

	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName 返回NFC表名
func (n *NFC) TableName() string {
	return TableName("nfc")
}

// BeforeCreate 在存储的时候DeviceID变成大写
func (n *NFC) BeforeCreate(_ *gorm.DB) (err error) {
	n.DeviceID = strings.ToUpper(n.DeviceID)
	return nil
}

// AfterFind 判断nfc内容类型,返回实际地址
func (n *NFC) AfterFind(_ *gorm.DB) (err error) {
	domainname := DefaultDomainName
	if config.Instance != nil && config.Instance.App != nil && config.Instance.App.DomainName != "" {
		domainname = config.Instance.App.DomainName
	}
	slog.Info("[NFC] AfterFind", "voice", n.Voice, "picture", n.Picture)
	if n.Voice != "" {
		deviceID, _, found := strings.Cut(n.Voice, "/")
		if found {
			n.Voice = fmt.Sprintf("%s/api/v1/file/%s/file_proxy?filename=%s", domainname, deviceID, n.Voice)
		}
	}
	if n.Picture != "" {
		deviceID, _, found := strings.Cut(n.Picture, "/")
		if found {
			n.Picture = fmt.Sprintf("%s/api/v1/file/%s/file_proxy?filename=%s", domainname, deviceID, n.Picture)
		}
	}
	return nil
}
