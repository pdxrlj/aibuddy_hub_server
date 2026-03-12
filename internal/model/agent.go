// Package model provides database models for the application.
package model

import "time"

// Agent represents an AI agent in the system.
type Agent struct {
	ID           int64  `gorm:"primaryKey;autoIncrement;column:id;"`
	UID          int64  `gorm:"column:uid;index;type:bigint;not null;default:0;comment:绑定用户id;"`
	AgentName    string `gorm:"column:agent_name;type:varchar(255);not null;comment:角色名称;"`
	DefaultUsage bool   `gorm:"column:default_usage;type:boolean;not null;default:false;comment:是否默认角色;"`

	RoleIntroduction string `gorm:"column:role_introduction;index;type:text;not null;comment:角色介绍;"`
	SystemPrompt     string `gorm:"column:system_prompt;type:text;index;not null;comment:系统提示词;"`

	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;not null;"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;not null;"`
}

// TableName returns the table name for Agent model.
func (Agent) TableName() string {
	return TableName("agent")
}
