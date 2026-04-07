// Package model 设备激活模型
package model

import (
	"time"

	"gorm.io/gorm"
)

// DeviceActivate 设备激活
type DeviceActivate struct {
	ID           int64     `gorm:"column:id;autoIncrement;primaryKey" json:"id"`
	DeviceID     string    `gorm:"column:device_id;type:varchar(255);comment:设备ID;index" json:"device_id"`
	ActivateTime LocalTime `gorm:"column:activate_time;comment:激活时间" json:"activate_time"`

	CreatedAt LocalTime `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt LocalTime `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
}

// TableName 表名
func (d *DeviceActivate) TableName() string {
	return TableName("device_activate")
}

// BeforeCreate 创建前钩子
func (d *DeviceActivate) BeforeCreate(_ *gorm.DB) error {
	d.ActivateTime = LocalTime(time.Now())
	d.CreatedAt = LocalTime(time.Now())
	d.UpdatedAt = LocalTime(time.Now())
	return nil
}

// BeforeUpdate 更新前钩子
func (d *DeviceActivate) BeforeUpdate(_ *gorm.DB) error {
	d.UpdatedAt = LocalTime(time.Now())
	return nil
}
