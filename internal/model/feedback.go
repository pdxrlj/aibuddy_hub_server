// Package model defines the data models for the application.
package model

import (
	"time"

	"gorm.io/datatypes"
)

// FeedbackType 反馈类型
type FeedbackType string

const (
	// FeedbackTypeFunctionException 功能异常
	FeedbackTypeFunctionException FeedbackType = "function_exception"

	// FeedbackTypeExperienceProblem 体验问题
	FeedbackTypeExperienceProblem FeedbackType = "experience_problem"

	// FeedbackTypeProductSuggestion 产品建议
	FeedbackTypeProductSuggestion FeedbackType = "product_suggestion"

	// FeedbackTypeOther 反馈类型其他
	FeedbackTypeOther FeedbackType = "other"
)

// Feedback 反馈
type Feedback struct {
	ID           int64        `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UID          int64        `gorm:"column:uid;not null" json:"uid"`
	Content      string       `gorm:"column:content;not null" json:"content"`
	FeedbackType FeedbackType `gorm:"column:feedback_type;not null" json:"feedback_type"`

	// 相关图片
	Images datatypes.JSON `gorm:"column:images;not null" json:"images"`

	CreatedAt time.Time `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null" json:"updated_at"`
}

// TableName 表名
func (f *Feedback) TableName() string {
	return TableName("feedback")
}
