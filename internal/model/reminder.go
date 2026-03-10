package model

import "time"

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

// RepeatType 重复类型
type RepeatType string

const (
	// RepeatTypeNone 不重复
	RepeatTypeNone RepeatType = "不重复"
	// RepeatTypeDaily 每天
	RepeatTypeDaily RepeatType = "每天"
	// RepeatTypeWeekly 每周
	RepeatTypeWeekly RepeatType = "每周"
	// RepeatTypeMonthly 每月
	RepeatTypeMonthly RepeatType = "每月"
	// RepeatTypeYearly 每年
	RepeatTypeYearly RepeatType = "每年"
)

func (r RepeatType) String() string {
	return string(r)
}

// IsValidReminderStatus 是否符合ReminderStatus
func IsValidReminderStatus(t ReminderStatus) bool {
	switch t {
	case ReminderStatusFailed, ReminderStatusCancelled, ReminderStatusPending, ReminderStatusCompleted:
		return true
	default:
		return false
	}
}

// IsValidRepeatType 是否符合IsValidRepeatType
func IsValidRepeatType(t RepeatType) bool {
	switch t {
	case RepeatTypeNone, RepeatTypeDaily, RepeatTypeWeekly, RepeatTypeMonthly, RepeatTypeYearly:
		return true
	default:
		return false
	}
}

// Reminder represents a reminder in the system.
type Reminder struct {
	ID int64 `gorm:"primaryKey;autoIncrement;column:id;"`
	// ReminderType int        `gorm:"column:reminder_type;type:int;not null;comment:提醒类型;"`
	RepeatType RepeatType `gorm:"column:repeat_type;type:varchar(50);not null;comment:重复类型;"`

	EntryID string `gorm:"column:entry_id;type:varchar(255);comment:任务ID;"`

	ReminderTitle   string `gorm:"column:reminder_title;type:varchar(255);not null;comment:提醒标题;"`
	ReminderContent string `gorm:"column:reminder_content;type:text;not null;comment:提醒内容;"`

	// 首次提醒
	ReminderTime time.Time `gorm:"column:reminder_time;type:timestamp;not null;comment:首次提醒时间;"`
	// 下一次提醒
	NextReminderTime time.Time `gorm:"column:next_reminder_time;type:timestamp;not null;comment:下一次提醒时间;"`

	Status ReminderStatus `gorm:"column:status;type:varchar(255);not null;default:待提醒;"`

	DeviceID string  `gorm:"column:device_id;index;type:varchar(255);not null;comment:设备ID;"`
	Device   *Device `gorm:"foreignKey:DeviceID;references:DeviceID;"`

	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;comment:创建时间;"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;comment:更新时间;"`
}

// TableName returns the table name for Reminder model.
func (r Reminder) TableName() string {
	return TableName("reminder")
}

// ComputedNextReminderTime 计算下一次提醒时间
func (r *Reminder) ComputedNextReminderTime() {
	switch r.RepeatType {
	case RepeatTypeNone:
		r.NextReminderTime = r.ReminderTime
	case RepeatTypeDaily:
		r.NextReminderTime = r.ReminderTime.AddDate(0, 0, 1)
	case RepeatTypeWeekly:
		r.NextReminderTime = r.ReminderTime.AddDate(0, 0, 7)
	case RepeatTypeMonthly:
		r.NextReminderTime = r.ReminderTime.AddDate(0, 1, 0)
	case RepeatTypeYearly:
		r.NextReminderTime = r.ReminderTime.AddDate(1, 0, 0)
	default:
		r.NextReminderTime = r.ReminderTime
	}
}
