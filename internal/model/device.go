package model

import "time"

// DeviceStatus represents the status of a device.
type DeviceStatus string

const (
	// DeviceStatusDisabled represents a disabled device.
	DeviceStatusDisabled DeviceStatus = "禁用"
	// DeviceStatusOffline represents an offline device.
	DeviceStatusOffline DeviceStatus = "离线"
	// DeviceStatusOnline represents an online device.
	DeviceStatusOnline DeviceStatus = "在线"
	// DeviceStatusUnknown represents an unknown device status.
	DeviceStatusUnknown DeviceStatus = "未知"
	// DeviceStatusAbnormal represents an abnormal device.
	DeviceStatusAbnormal DeviceStatus = "异常"
	// DeviceStatusFault represents a faulty device.
	DeviceStatusFault DeviceStatus = "故障"
)

// String returns the string representation of DeviceStatus.
func (s DeviceStatus) String() string {
	return string(s)
}

// Device represents a device in the system.
type Device struct {
	ID          int64     `gorm:"primaryKey;autoIncrement;column:id;"`
	DeviceID    string    `gorm:"column:device_id;index;type:varchar(255);not null;uniqueIndex;comment:设备ID;"`
	LastLoginAt time.Time `gorm:"column:last_login_at;type:timestamp;comment:最后登录时间;"`

	Status DeviceStatus `gorm:"column:status;type:varchar(50);not null;default:未知;comment:状态:未知;"`

	AgentID int64 `gorm:"column:agent_id;index;type:bigint;not null;comment:角色ID;"`

	Agent *Agent `gorm:"foreignKey:AgentID;references:ID;"`

	User *User `gorm:"foreignKey:ID;references:DeviceID;"`

	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;comment:创建时间;"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;comment:更新时间;"`
}

// TableName returns the table name for Device model.
func (Device) TableName() string {
	return TableName("device")
}
