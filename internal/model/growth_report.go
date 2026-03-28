// Package model provides the model for the application.
package model

import (
	"encoding/json"
	"errors"
	"strings"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// GrowthReport 群组报告
type GrowthReport struct {
	ID          int64     `gorm:"column:id;type:bigint(20);primaryKey;autoIncrement" json:"id"`
	DeviceID    string    `gorm:"column:device_id;type:varchar(255);index;not null" json:"device_id"`
	StartTime   LocalTime `gorm:"column:start_time;type:timestamp;not null" json:"start_time"`
	EndTime     LocalTime `gorm:"column:end_time;type:timestamp;not null" json:"end_time"`
	SummaryText string    `gorm:"column:summary_text;type:text;not null" json:"summary_text"`

	StatusCards          datatypes.JSON `gorm:"column:status_cards;type:json;not null" json:"status_cards"`
	InteractionSummary   datatypes.JSON `gorm:"column:interaction_summary;type:json;not null" json:"interaction_summary"`
	SocialSummary        datatypes.JSON `gorm:"column:social_summary;type:json;not null" json:"social_summary"`
	MemoryCapsuleSummary datatypes.JSON `gorm:"column:memory_capsule_summary;type:json;not null" json:"memory_capsule_summary"`
	ChildPortrait        datatypes.JSON `gorm:"column:child_portrait;type:json;not null" json:"child_portrait"`
	KeyMoments           datatypes.JSON `gorm:"column:key_moments;type:json;not null" json:"key_moments"`
	EmotionTrend         datatypes.JSON `gorm:"column:emotion_trend;type:json;not null" json:"emotion_trend"`
	AudioSummary         datatypes.JSON `gorm:"column:audio_summary;type:json;not null" json:"audio_summary"`
	PomodoroSummary      datatypes.JSON `gorm:"column:pomodoro_summary;type:json;not null" json:"pomodoro_summary"`
	SafetyAlert          datatypes.JSON `gorm:"column:safety_alert;type:json;not null" json:"safety_alert"`
	NextWeekSuggestions  datatypes.JSON `gorm:"column:next_week_suggestions;type:json;not null" json:"next_week_suggestions"`
	ParentScripts        datatypes.JSON `gorm:"column:parent_scripts;type:json;not null" json:"parent_scripts"`
	ClosingText          string         `gorm:"column:closing_text;type:text;not null" json:"closing_text"`

	CreatedAt LocalTime `gorm:"column:created_at;type:timestamp;not null" json:"created_at"`
	UpdatedAt LocalTime `gorm:"column:updated_at;type:timestamp;not null" json:"updated_at"`
}

// TableName 返回群组报告表名
func (g *GrowthReport) TableName() string {
	return TableName("growth_report")
}

// String 返回群组报告字符串
func (g *GrowthReport) String() ([]byte, error) {
	if g == nil {
		return nil, errors.New("growth report is nil")
	}
	jsonData, err := json.Marshal(g)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

// MustString 返回群组报告字符串，如果失败则返回空字符串
func (g *GrowthReport) MustString() []byte {
	jsonData, err := g.String()
	if err != nil {
		return []byte{}
	}
	return jsonData
}

// BeforeCreate 在插入之前
func (g *GrowthReport) BeforeCreate(_ *gorm.DB) (err error) {
	if g.DeviceID != "" {
		g.DeviceID = strings.ToUpper(g.DeviceID)
	}
	return nil
}
