// Package model provides the model for the application.
package model

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// GrowthReport 群组报告
type GrowthReport struct {
	ID          int64     `gorm:"column:id;type:bigint(20);primaryKey;autoIncrement"`
	DeviceID    string    `gorm:"column:device_id;type:varchar(255);index;not null"`
	StartTime   time.Time `gorm:"column:start_time;type:timestamp;not null"`
	EndTime     time.Time `gorm:"column:end_time;type:timestamp;not null"`
	SummaryText string    `gorm:"column:summary_text;type:text;not null"`

	StatusCards          datatypes.JSON `gorm:"column:status_cards;type:json;not null"`
	InteractionSummary   datatypes.JSON `gorm:"column:interaction_summary;type:json;not null"`
	SocialSummary        datatypes.JSON `gorm:"column:social_summary;type:json;not null"`
	MemoryCapsuleSummary datatypes.JSON `gorm:"column:memory_capsule_summary;type:json;not null"`
	ChildPortrait        datatypes.JSON `gorm:"column:child_portrait;type:json;not null"`
	KeyMoments           datatypes.JSON `gorm:"column:key_moments;type:json;not null"`
	EmotionTrend         datatypes.JSON `gorm:"column:emotion_trend;type:json;not null"`
	AudioSummary         datatypes.JSON `gorm:"column:audio_summary;type:json;not null"`
	PomodoroSummary      datatypes.JSON `gorm:"column:pomodoro_summary;type:json;not null"`
	SafetyAlert          datatypes.JSON `gorm:"column:safety_alert;type:json;not null"`
	NextWeekSuggestions  datatypes.JSON `gorm:"column:next_week_suggestions;type:json;not null"`
	ParentScripts        datatypes.JSON `gorm:"column:parent_scripts;type:json;not null"`
	ClosingText          string         `gorm:"column:closing_text;type:text;not null"`

	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;not null"`
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
