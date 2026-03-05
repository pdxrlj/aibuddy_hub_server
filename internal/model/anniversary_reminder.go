// Package model provides model definitions.
package model

import "time"

// AnniversaryType 纪念日类型
type AnniversaryType string

const (
	// AnniversaryTypeBirthday 生日
	AnniversaryTypeBirthday AnniversaryType = "生日"
	// AnniversaryTypeGraduation 毕业
	AnniversaryTypeGraduation AnniversaryType = "毕业"
	// AnniversaryTypeMeet 相识
	AnniversaryTypeMeet AnniversaryType = "相识"
	// AnniversaryTypeMarriage 结婚
	AnniversaryTypeMarriage AnniversaryType = "结婚"
	// AnniversaryTypeLove 恋爱
	AnniversaryTypeLove AnniversaryType = "恋爱"
	// AnniversaryTypeWork 工作
	AnniversaryTypeWork AnniversaryType = "工作"
	// AnniversaryTypeOther 其他
	AnniversaryTypeOther AnniversaryType = "其他"
)

func (r AnniversaryType) String() string {
	return string(r)
}

// ReminderWay 提醒方式
type ReminderWay string

const (
	// ReminderWayOnDay 当天
	ReminderWayOnDay ReminderWay = "当天"
	// ReminderWayBeforeOneDay 提前一天
	ReminderWayBeforeOneDay ReminderWay = "提前一天"
	// ReminderWayBeforeThreeDays 提前三天
	ReminderWayBeforeThreeDays ReminderWay = "提前三天"
	// ReminderWayBeforeOneWeek 提前一周
	ReminderWayBeforeOneWeek ReminderWay = "提前一周"
)

func (r ReminderWay) String() string {
	return string(r)
}

// AnniversaryReminder 纪念日提醒
type AnniversaryReminder struct {
	ID int64 `gorm:"primaryKey;autoIncrement;column:id;"`

	EntryID         string          `gorm:"column:entry_id;type:varchar(255);not null;comment:任务ID;"`
	AnniversaryType AnniversaryType `gorm:"column:anniversary_type;type:varchar(50);not null;comment:纪念日类型;"`

	ReminderUsername string    `gorm:"column:reminder_username;type:varchar(50);not null;comment:提醒人用户名;"`
	ReminderUserSex  string    `gorm:"column:reminder_user_sex;type:varchar(50);not null;comment:提醒人性别;男/女;"`
	AnniversaryTime  time.Time `gorm:"column:anniversary_time;type:timestamp;not null;comment:纪念日时间;"`

	// 提醒方式
	ReminderWay ReminderWay `gorm:"column:reminder_way;type:varchar(50);not null;comment:提醒方式;"`

	// 备注
	Remarks string `gorm:"column:remarks;type:text;not null;comment:备注;"`

	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;not null;comment:创建时间;"`

	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;not null;comment:更新时间;"`
}

// TableName 返回表名
func (a *AnniversaryReminder) TableName() string {
	return TableName("anniversary_reminder")
}
