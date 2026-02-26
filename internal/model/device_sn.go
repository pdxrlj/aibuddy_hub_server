// Package model 设备SN模型
package model

import "time"

// DeviceSN 设备SN
type DeviceSN struct {
	ID        int64     `gorm:"column:id;type:bigint(20);primaryKey;autoIncrement"`
	SN        string    `gorm:"column:sn;type:varchar(45);not null;unique"`
	DeviceID  string    `gorm:"column:device_id;type:varchar(45);unique"`
	IsValid   bool      `gorm:"column:is_valid;type:boolean;default:true"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP"`
}

// TableName 表名
func (DeviceSN) TableName() string {
	return TableName("device_sn")
}
