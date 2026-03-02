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
	ID        int64  `gorm:"primaryKey;autoIncrement;column:id;"`
	DeviceID  string `gorm:"column:device_id;index;type:varchar(255);not null;uniqueIndex;comment:设备ID;"`
	ICCID     string `gorm:"column:iccid;index;type:varchar(20);not null;uniqueIndex;comment:ICCID;comment:手机卡ICCID;"`
	BoardType string `gorm:"column:board_type;type:varchar(255);not null;comment:板子类型;"`
	Version   string `gorm:"column:version;type:varchar(255);not null;comment:板子版本;"`

	UID          int64     `gorm:"column:uid;index;type:bigint;not null;default:0;comment:绑定用户id;"`
	LastActiveAt time.Time `gorm:"column:last_active_at;type:timestamp;comment:最后活跃时间;"`

	Status  DeviceStatus `gorm:"column:status;type:varchar(50);not null;default:未知;comment:状态:未知;"`
	IsAdmin bool         `gorm:"column:is_admin;type:boolean;not null;default:false;comment:是否管理员设备,首次绑定的用户是管理员;"`
	AgentID int64        `gorm:"column:agent_id;index;type:bigint;not null;comment:角色ID;"`

	Agent      *Agent      `gorm:"foreignKey:AgentID;references:ID;"`
	DeviceInfo *DeviceInfo `gorm:"foreignKey:DeviceID;references:DeviceID;"`

	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;comment:创建时间;"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;comment:更新时间;"`
}

// TableName returns the table name for Device model.
func (Device) TableName() string {
	return TableName("device")
}
