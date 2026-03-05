// Package baidu 百度云API客户端角色切换
package baidu

import (
	"fmt"
	"net/url"

	"aibuddy/pkg/config"
)

// SwitchRole 角色切换API
type SwitchRole struct {
	client *Client
}

// NewSwitchRole 创建角色切换客户端
func NewSwitchRole() *SwitchRole {
	return &SwitchRole{client: NewClient()}
}

// SwitchRoleRequest 切换角色请求
type SwitchRoleRequest struct {
	AppID             string // 大模型互动应用ID
	AiAgentInstanceID uint64 `json:"ai_agent_instance_id"` // 大模型互动实例ID
	SceneRole         string `json:"scene_role"`           // 角色名称（需已在控制台创建）
	TTS               string `json:"tts,omitempty"`        // TTS配置，取值等价于创建实例接口的tts_url
	TTSSayHi          string `json:"tts_sayhi,omitempty"`  // 切换音色后的招呼语
}

// SwitchSceneRole 切换大模型互动实例的角色（音色）
func (s *SwitchRole) SwitchSceneRole(req *SwitchRoleRequest) error {
	appID := req.AppID
	if appID == "" {
		appID = config.Instance.Baidu.AppID
	}

	path := "/api/v1/aiagent/instances/operation"
	query := url.Values{}
	query.Set("switchSceneRole", "")

	body := map[string]any{
		"app_id":               appID,
		"ai_agent_instance_id": req.AiAgentInstanceID,
		"scene_role":           req.SceneRole,
	}
	if req.TTS != "" {
		body["tts"] = req.TTS
	}
	if req.TTSSayHi != "" {
		body["tts_sayhi"] = req.TTSSayHi
	}

	if err := s.client.Request("PUT", path, query, body, nil); err != nil {
		return fmt.Errorf("切换角色失败: %w", err)
	}

	return nil
}
