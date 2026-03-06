// Package repository 聊天记录仓库
package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"time"

	"gorm.io/gen"
)

// ChatDialogueRepository 聊天记录仓库
type ChatDialogueRepository struct{}

// NewChatDialogueRepository 创建聊天记录仓库
func NewChatDialogueRepository() *ChatDialogueRepository {
	return &ChatDialogueRepository{}
}

// GetChatDialogue 获取聊天记录
// startDate 开始日期
// endDate 结束日期
// roleAgentName 角色代理名称,nil 为查询所有角色的聊天记录
// 查询聊天记录获取对应角色代理的聊天记录
// 返回聊天记录
func (r *ChatDialogueRepository) GetChatDialogue(startDate, endDate time.Time, roleAgentName ...string) ([]*model.ChatDialogue, error) {
	dialogue, err := query.ChatDialogue.Scopes(func(db gen.Dao) gen.Dao {
		if len(roleAgentName) > 0 {
			return db.Where(query.ChatDialogue.AgentName.In(roleAgentName...))
		}
		return db
	}).Where(
		query.ChatDialogue.CreatedAt.Between(startDate, endDate),
	).Find()

	return dialogue, err
}
