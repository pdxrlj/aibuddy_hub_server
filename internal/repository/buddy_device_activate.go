// Package repository 提供数据库操作相关功能
package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"

	"gorm.io/gorm"
)

// BuddyDeviceActivateRepo 设备激活仓库
type BuddyDeviceActivateRepo struct {
}

// NewBuddyDeviceActivateRepo 创建设备激活仓库实例
func NewBuddyDeviceActivateRepo() *BuddyDeviceActivateRepo {
	return &BuddyDeviceActivateRepo{}
}

// IsActivatedDevice 判断设备是否已经激活了，如果没有激活就送一个月VIP
func (n *BuddyDeviceActivateRepo) IsActivatedDevice(deviceID string) bool {
	_, err := query.DeviceActivate.Where(query.DeviceActivate.DeviceID.Eq(deviceID)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false
		}
		return false
	}
	return true
}

// CreateDeviceActivate 创建设备激活
func (n *BuddyDeviceActivateRepo) CreateDeviceActivate(deviceID string, tx ...*query.Query) error {
	db := query.Q
	if len(tx) > 0 {
		db = tx[0]
	}
	return db.DeviceActivate.Create(
		&model.DeviceActivate{
			DeviceID: deviceID,
		},
	)
}

// Transaction 执行事务
func (n *BuddyDeviceActivateRepo) Transaction(fn func(tx *query.Query) error) error {
	return query.Q.Transaction(fn)
}
