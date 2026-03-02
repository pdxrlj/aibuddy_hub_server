// Package role 角色请求体
package role

import "time"

// ListRequest 角色列表请求
type ListRequest struct {
	Page int `json:"page" form:"page" param:"page" query:"page" validate:"gte=1" default:"1"`
	Size int `json:"size" form:"size" param:"size" query:"size" validate:"gte=1" default:"10"`
}

// ListResponse 列表响应体
type ListResponse struct {
	Total int             `json:"total"`
	Page  int             `json:"page"`
	Size  int             `json:"size"`
	Roles []RolesResponse `json:"role,omitempty"`
}

// RolesResponse 角色响应体
type RolesResponse struct {
	ID        int64  `json:"id"`
	UID       int64  `json:"uid"`
	AgentName string `json:"agent_name"`

	RoleIntroduction string `json:"role_introduction"`
	SystemPrompt     string `json:"system_prompt"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ChangeRquest 切换角色请求
type ChangeRquest struct {
	DeviceID string `json:"device_id" validate:"required" msg:"required:设备ID不能为空"`
	RoleID   int64  `json:"role_id" validate:"required" msg:"required:角色ID不能为空"`
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
