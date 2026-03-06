// Package agent 代理处理器
package agent

// RoleChatAgentRequest 角色聊天代理请求
type RoleChatAgentRequest struct {
	StartDate string `json:"start_date" form:"start_date" validate:"required,date" msg:"start_date不能为空|date:start_date格式错误"`
	EndDate   string `json:"end_date" form:"end_date" validate:"required,date" msg:"end_date不能为空|date:end_date格式错误"`

	RoleAgentName string `json:"role_agent_name" form:"role_agent_name" validate:"required" msg:"role_agent_name不能为空"`
}
