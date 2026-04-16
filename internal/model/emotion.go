// Package model 情绪预警触发记录
package model

import (
	"strings"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// EmotionStatus 情绪预警状态
type EmotionStatus string

const (
	// EmotionStatusPending 待处理
	EmotionStatusPending EmotionStatus = "pending"
	// EmotionStatusRead 已读
	EmotionStatusRead EmotionStatus = "read"
	// EmotionStatusHandled 已处理
	EmotionStatusHandled EmotionStatus = "handled"
)

// Emotion 情绪预警触发记录
type Emotion struct {
	ID         int64  `gorm:"primaryKey;autoIncrement;column:id;" json:"id"`
	DeviceID   string `gorm:"column:device_id;index;type:varchar(32);not null;comment:设备ID;" json:"device_id"`
	DialogueID int64  `gorm:"column:dialogue_id;index;comment:关联的对话记录ID;" json:"dialogue_id"`

	// 预警基本信息
	TriggerWarning bool           `gorm:"column:trigger_warning;type:boolean;not null;default:false;comment:是否触发预警;" json:"trigger_warning"`
	WarningLevel   string         `gorm:"column:warning_level;type:varchar(10);comment:预警等级(低/中/高);" json:"warning_level"`
	WarningTypes   datatypes.JSON `gorm:"column:warning_types;type:varchar(255);comment:预警类型,逗号分隔;" json:"warning_types"`
	Confidence     float64        `gorm:"column:confidence;type:decimal(3,2);comment:置信度(0-1);" json:"confidence"`

	// 预警原因
	WarningReason datatypes.JSON `gorm:"column:warning_reason;type:json;comment:预警原因;" json:"warning_reason"`

	// 证据
	Evidence datatypes.JSON `gorm:"column:evidence;type:json;comment:证据;" json:"evidence"`

	// 家长建议
	ParentSuggestions datatypes.JSON `gorm:"column:parent_suggestions;type:json;comment:家长建议;" json:"parent_suggestions"`

	// 是否需要人工跟进
	NeedManualFollowup bool `gorm:"column:need_manual_followup;type:boolean;default:false;comment:是否需要人工跟进;" json:"need_manual_followup"`

	PrivacyRisk   datatypes.JSON `gorm:"column:privacy_risk;type:json;comment:隐私风险;" json:"privacy_risk"`
	ScamRisk      datatypes.JSON `gorm:"column:scam_risk;type:json;comment:诈骗风险;" json:"scam_risk"`
	EmotionalRisk datatypes.JSON `gorm:"column:emotional_risk;type:json;comment:情绪风险;" json:"emotional_risk"`

	// 整体评估
	OverallAssessment string `gorm:"column:overall_assessment;type:text;comment:整体评估;" json:"overall_assessment"`

	// 阅读状态
	Read bool `gorm:"column:read;type:boolean;default:false;comment:是否已读;" json:"read"`

	CreatedAt LocalTime `gorm:"column:created_at;type:timestamp;not null;comment:创建时间;" json:"created_at"`
	UpdatedAt LocalTime `gorm:"column:updated_at;type:timestamp;not null;comment:更新时间;" json:"updated_at"`
}

// TableName 表名
func (Emotion) TableName() string {
	return TableName("emotion")
}

// BeforeCreate 在插入之前
func (e *Emotion) BeforeCreate(_ *gorm.DB) (err error) {
	if e.DeviceID != "" {
		e.DeviceID = strings.ToUpper(e.DeviceID)
	}
	return nil
}
