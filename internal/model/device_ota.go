// Package model provides the device ota model.
package model

import (
	"time"
)

// DeviceOta 设备 OTA 信息
type DeviceOta struct {
	ID       int    `gorm:"column:id;type:int(11);primaryKey;autoIncrement;"`
	DeviceID string `gorm:"column:device_id;uniqueIndex;type:varchar(255);index;not null;"`

	Device *Device `gorm:"foreignKey:DeviceID;references:DeviceID;"`

	Version     string    `gorm:"column:version;type:varchar(255);not null;"`
	OtaURL      string    `gorm:"column:ota_url;type:varchar(255);not null;"`
	ModelURL    string    `gorm:"column:model_url;type:varchar(255);not null;"`
	ResourceURL string    `gorm:"column:resource_url;type:varchar(255);not null;"`
	ForceUpdate bool      `gorm:"column:force_update;type:boolean;not null;"`
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamp;not null;"`
	UpdatedAt   time.Time `gorm:"column:updated_at;type:timestamp;not null;"`
}

// TableName 表名
func (DeviceOta) TableName() string {
	return TableName("device_ota")
}
