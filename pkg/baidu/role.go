package baidu

import (
	"fmt"

	"aibuddy/pkg/config"
)

// Role 角色API
type Role struct {
	client *Client
}

// NewRole 创建角色客户端
func NewRole() *Role {
	return &Role{client: NewClient()}
}

// SystemRole 系统角色
type SystemRole struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// SystemRolesResponse 系统角色列表响应
type SystemRolesResponse struct {
	Data []SystemRole `json:"data"`
}

// RoleInfo 角色信息
type RoleInfo struct {
	ID           int64  `json:"id,omitempty"`
	Name         string `json:"name,omitempty"`
	Description  string `json:"description,omitempty"`
	SystemRoleID int64  `json:"systemRoleId,omitempty"`
	Prompt       string `json:"prompt,omitempty"`
	Model        string `json:"model,omitempty"`
	DefaultUsage bool   `json:"defaultUsage,omitempty"`
}

// LLMConfig 大模型配置
type LLMConfig struct {
	Roles []RoleInfo `json:"roles"`
}

// AppDetailResponse 应用详情响应
type AppDetailResponse struct {
	LLM LLMConfig `json:"llm"`
}

// GetSystemRoles 查询系统角色列表
func (r *Role) GetSystemRoles() (*SystemRolesResponse, error) {
	path := "/api/v1/scene_roles"
	var result SystemRolesResponse

	if err := r.client.Request("GET", path, nil, nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// RoleList 获取角色列表
func (r *Role) RoleList(appID string) (*AppDetailResponse, error) {
	if appID == "" {
		appID = config.Instance.Baidu.AppID
	}

	path := fmt.Sprintf("/api/v1/apps/%s", appID)
	var result AppDetailResponse

	if err := r.client.Request("GET", path, nil, nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// BindFunctionTemplateRequest 绑定Function模板请求
type BindFunctionTemplateRequest struct {
	Type               string `json:"type"`
	FunctionTemplateID int64  `json:"functionTemplateId"`
}

// BindFunctionTemplate 绑定Function模板
func (r *Role) BindFunctionTemplate(appID string, req *BindFunctionTemplateRequest) error {
	if appID == "" {
		appID = config.Instance.Baidu.AppID
	}

	path := fmt.Sprintf("/api/v1/apps/%s/llm", appID)
	return r.client.Request("PUT", path, nil, req, nil)
}

// RoleAddItem 新增角色项
type RoleAddItem struct {
	SystemRoleID int64  `json:"systemRoleId,omitempty"`
	Name         string `json:"name,omitempty"`
	Description  string `json:"description,omitempty"`
	Prompt       string `json:"prompt,omitempty"`
	Model        string `json:"model,omitempty"`
	DefaultUsage bool   `json:"defaultUsage,omitempty"`
}

// RoleUpdateItem 更新角色项
type RoleUpdateItem struct {
	ID           int64  `json:"id"`
	Name         string `json:"name,omitempty"`
	Description  string `json:"description,omitempty"`
	Prompt       string `json:"prompt,omitempty"`
	Model        string `json:"model,omitempty"`
	DefaultUsage bool   `json:"defaultUsage,omitempty"`
}

// RoleOperation 角色操作
type RoleOperation struct {
	AddItems    []RoleAddItem    `json:"addItems,omitempty"`
	UpdateItems []RoleUpdateItem `json:"updateItems,omitempty"`
	RemoveItems []int64          `json:"removeItems,omitempty"`
}

// ManageRolesRequest 角色管理请求
type ManageRolesRequest struct {
	Type string         `json:"type"`
	Role *RoleOperation `json:"role"`
}

// ManageRoles 角色的增、改、删
func (r *Role) ManageRoles(appID string, req *ManageRolesRequest) error {
	if appID == "" {
		appID = config.Instance.Baidu.AppID
	}

	path := fmt.Sprintf("/api/v1/apps/%s/llm", appID)
	return r.client.Request("PUT", path, nil, req, nil)
}
