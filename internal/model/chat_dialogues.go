// Package model 聊天对话
package model

import "time"

// ChatDialogue 聊天对话记录（问答对）
type ChatDialogue struct {
	ID           int64     `gorm:"primaryKey;autoIncrement;column:id;"`
	DeviceID     string    `gorm:"column:device_id;index;type:varchar(32);not null;comment:设备ID;"`
	Question     string    `gorm:"column:question;type:text;not null;comment:问题内容;"`
	QuestionTime time.Time `gorm:"column:question_time;type:timestamp;not null;comment:问题时间;"`
	Answer       string    `gorm:"column:answer;type:text;not null;comment:回答内容;"`
	AnswerTime   time.Time `gorm:"column:answer_time;type:timestamp;not null;comment:回答时间;"`

	AgentName string `gorm:"column:agent_name;type:varchar(255);comment:角色名称;"`

	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;not null;comment:创建时间;"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;not null;comment:更新时间;"`
}

// TableName 表名
func (ChatDialogue) TableName() string {
	return TableName("chat_dialogue")
}
