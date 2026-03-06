// Package agent 代理处理器
package agent

import "aibuddy/pkg/ahttp"

// Handler 代理处理器
type Handler struct{}

// NewHandler 创建代理处理器
func NewHandler() *Handler {
	return &Handler{}
}

// RoleChatAgent 角色聊天代理
func (h *Handler) RoleChatAgent(state *ahttp.State, req *RoleChatAgentRequest) error {
	_ = state
	_ = req
	return nil
}
