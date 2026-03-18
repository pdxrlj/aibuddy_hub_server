// Package model provides the model for the NFC card
package model

import "time"

// NFCStatus NFC状态
type NFCStatus string

const (
	// NFCPending 制作中
	NFCPending NFCStatus = "制作中"

	// NFCPaid 制作完成
	NFCPaid NFCStatus = "制作完成"
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

	Title   string `gorm:"column:title;type:varchar(255);not null" json:"title"`
	Content string `gorm:"column:content;type:text;not null" json:"content"`

	Status NFCStatus `gorm:"column:status;type:varchar(255);not null;default:'制作中'" json:"status"`

	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName 返回NFC表名
func (n *NFC) TableName() string {
	return TableName("nfc")
}
