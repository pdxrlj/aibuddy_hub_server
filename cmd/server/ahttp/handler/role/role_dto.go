// Package role 角色请求体
package role

import "time"

// ListRequest 角色列表请求
type ListRequest struct {
	Page int `json:"page" form:"page" param:"page" query:"page" validate:"gte=1" default:"1"`
	Size int `json:"size" form:"size" param:"size" query:"size" validate:"gte=1" default:"10"`
}

// RolesResponse 角色响应体
type RolesResponse struct {
	ID               int64  `json:"id"`
	AgentName        string `json:"agent_name"`
	DefaultUsage     bool   `json:"default_usage"`
	RoleIntroduction string `json:"role_introduction"`
}

// ChangeRquest 切换角色请求
type ChangeRquest struct {
	DeviceID string `json:"device_id" validate:"required" msg:"required:设备ID不能为空"`
	RoleName string `json:"role_name" validate:"required" msg:"required:角色不能为空"`
}

// InfoRequest 角色信息请求数据
type InfoRequest struct {
	RoleID int64 `json:"role_id" form:"role_id" param:"role_id" query:"role_id" validate:"required" msg:"required:角色ID不能为空"`
}

// InfoResponse 角色信息响应
type InfoResponse struct {
	ID               int64  `json:"id"`
	UID              int64  `json:"uid"`
	AgentName        string `json:"agent_name"`
	RoleIntroduction string `json:"role_introduction"`
	SystemPrompt     string `json:"system_prompt"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

// GetChatAnalysisRequest 获取聊天分析请求
type GetChatAnalysisRequest struct {
	DeviceID  string `json:"device_id" param:"device_id"  validate:"required,aimac" msg:"required:设备ID不能为空|aimac:不是有效的设备号"`
	AgentName string `json:"agent_name" query:"agent_name"  validate:"required,min=1,max=255" msg:"required:角色名称不能为空|min:1|max:255:角色名称不能超过255个字符"`
}

// RefreshChatAnalysisRequest 聊天分析请求
type RefreshChatAnalysisRequest struct {
	DeviceID  string    `json:"device_id" param:"device_id"  validate:"required,aimac" msg:"required:设备ID不能为空|aimac:不是有效的设备号"`
	StartTime time.Time `json:"start_time" query:"start_time"`
	EndTime   time.Time `json:"end_time" query:"end_time"`

	AgentName string `json:"agent_name" query:"agent_name"  validate:"required,min=1,max=255" msg:"required:角色名称不能为空|min:1|max:255:角色名称不能超过255个字符"`
}

// NormalizeTime 规范化时间，如果为空则设置为最近7天
func (r *RefreshChatAnalysisRequest) NormalizeTime() {
	now := time.Now()
	if r.EndTime.IsZero() {
		r.EndTime = now
	}
	if r.StartTime.IsZero() {
		r.StartTime = now.AddDate(0, 0, -7)
	}
}

// ChatAnalysisResponse 聊天分析响应
type ChatAnalysisResponse struct {
	ConversationAnalysis string `json:"conversation_analysis"`
	EmotionAnalysis      string `json:"emotion_analysis"`
}
