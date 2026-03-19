// Package model 用户代理模型
package model

import (
	"strings"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// UserAgent 用户代理
type UserAgent struct {
	ID       int    `gorm:"column:id;type:int(11);autoIncrement;primaryKey"`
	DeviceID string `gorm:"column:device_id;type:varchar(255);index;not null;comment:设备ID"`

	AgentName string `gorm:"column:agent_name;type:varchar(255);not null;comment:角色名称"`

	ConversationAnalysis datatypes.JSON `gorm:"column:conversation_analysis;type:json;not null;comment:对话分析"`
	EmotionAnalysis      datatypes.JSON `gorm:"column:emotion_analysis;type:json;not null;comment:情绪分析"`

	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;not null;default:current_timestamp;comment:创建时间"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;not null;default:current_timestamp;comment:更新时间"`
}

// TableName 表名
func (UserAgent) TableName() string {
	return TableName("user_agents")
}

// BeforeCreate 在插入之前
func (u *UserAgent) BeforeCreate(_ *gorm.DB) (err error) {
	if u.DeviceID != "" {
		u.DeviceID = strings.ToUpper(u.DeviceID)
	}
	return nil
}
