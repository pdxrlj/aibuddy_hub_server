// Package model is the model for the pomodoro clock
package model

import "time"

// PomodoroClock is the model for the pomodoro clock
type PomodoroClock struct {
	ID            int    `gorm:"column:id;type:int;primary_key;autoIncrement;comment:ID;" json:"id"`
	DeviceID      string `gorm:"column:device_id;type:varchar(36);comment:设备ID;" json:"device_id"`
	TotalDuration int    `gorm:"column:total_duration;type:int;comment:总时长;" json:"total_duration"`

	StudyDuration       int `gorm:"column:study_duration;type:int;comment:学习时长;" json:"study_duration"`
	DistractionDuration int `gorm:"column:distraction_duration;type:int;comment:分心时长;" json:"distraction_duration"`

	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;comment:创建时间;autoCreateTime;" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;comment:更新时间;" json:"updated_at"`
}

// TableName returns the table name for the pomodoro clock
func (PomodoroClock) TableName() string {
	return TableName("pomodoro_clock")
}
