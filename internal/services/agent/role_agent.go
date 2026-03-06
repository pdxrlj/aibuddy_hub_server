// Package agent 角色代理服务
package agent

import (
	"aibuddy/internal/model"
	"encoding/json"
)

// RoleAgentService 角色代理服务
type RoleAgentService struct{}

// NewRoleAgentService 创建角色代理服务
func NewRoleAgentService() *RoleAgentService {
	return &RoleAgentService{}
}

// RoleChatAgent 角色聊天代理
// startDate 开始日期
// endDate 结束日期
// roleAgentName 角色代理名称
// 查询聊天记录获取对应角色代理的聊天记录
// 生成用户角色使用的报告
func (s *RoleAgentService) RoleChatAgent(startDate, endDate, roleAgentName string) error {
	_ = startDate
	_ = endDate
	_ = roleAgentName
	return nil
}

// FormatChatToTemplate 格式化聊天记录为模板
type FormatChatToTemplate struct {
	Query  string `json:"query"`
	Answer string `json:"answer"`
}

// FormatChatToTemplate 格式化聊天记录为模板
func (s *RoleAgentService) FormatChatToTemplate(dialogue []*model.ChatDialogue) (string, error) {
	formatChatToTemplate := make([]*FormatChatToTemplate, 0, len(dialogue))

	for _, d := range dialogue {
		formatChatToTemplate = append(formatChatToTemplate, &FormatChatToTemplate{
			Query:  d.Question,
			Answer: d.Answer,
		})
	}

	jsonData, err := json.Marshal(formatChatToTemplate)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}
