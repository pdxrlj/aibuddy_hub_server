package model

import "time"

// ReminderType represents the type of reminder.
type ReminderType string

const (
	// ReminderTypeBirthday represents a birthday reminder.
	ReminderTypeBirthday ReminderType = "生日"
	// ReminderTypeGraduation represents a graduation reminder.
	ReminderTypeGraduation ReminderType = "毕业"
	// ReminderTypeMeet represents a meet reminder.
	ReminderTypeMeet ReminderType = "相识"
	// ReminderTypeMarriage represents a marriage reminder.
	ReminderTypeMarriage ReminderType = "结婚"
	// ReminderTypeLove represents a love reminder.
	ReminderTypeLove ReminderType = "恋爱"
	// ReminderTypeWork represents a work reminder.
	ReminderTypeWork ReminderType = "工作"
	// ReminderTypeOther represents other reminder types.
	ReminderTypeOther ReminderType = "其他"
)

func (r ReminderType) String() string {
	return string(r)
}

// ReminderStatus represents the status of a reminder.
type ReminderStatus string

const (
	// ReminderStatusPending represents a pending reminder.
	ReminderStatusPending ReminderStatus = "待提醒"
	// ReminderStatusCompleted represents a completed reminder.
	ReminderStatusCompleted ReminderStatus = "已完成"
	// ReminderStatusCancelled represents a cancelled reminder.
	ReminderStatusCancelled ReminderStatus = "已取消"
	// ReminderStatusFailed represents a failed reminder.
	ReminderStatusFailed ReminderStatus = "失败"
)

func (r ReminderStatus) String() string {
	return string(r)
}

// Reminder represents a reminder in the system.
type Reminder struct {
	ID              int64        `gorm:"primaryKey;autoIncrement;column:id;"`
	ReminderType    ReminderType `gorm:"column:reminder_type;type:varchar(50);not null;comment:提醒类型;"`
	ReminderTime    time.Time    `gorm:"column:reminder_time;type:timestamp;not null;comment:提醒时间;"`
	ReminderContent string       `gorm:"column:reminder_content;type:text;not null;comment:提醒内容;"`
	DeviceID        string       `gorm:"column:device_id;index;type:varchar(255);not null;comment:设备ID;"`

	// 提醒类型 按照cron表达式 0 0 0 * * * 每天0点0分0秒提醒
	CronExpression string `gorm:"column:cron_expression;type:varchar(255);not null;comment:cron表达式;"`

	Status ReminderStatus `gorm:"column:status;type:varchar(255);not null;default:待提醒;"`

	Device *Device `gorm:"foreignKey:DeviceID;references:DeviceID;"`

	ReminderDeviceID int64     `gorm:"column:reminder_device_id;type:bigint;not null;comment:提醒设备ID;"`
	CreatedAt        time.Time `gorm:"column:created_at;type:timestamp;comment:创建时间;"`
	UpdatedAt        time.Time `gorm:"column:updated_at;type:timestamp;comment:更新时间;"`
}

// TableName returns the table name for Reminder model.
func (r Reminder) TableName() string {
	return TableName("reminder")
}
