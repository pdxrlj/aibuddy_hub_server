// Package baidu 百度云API客户端记忆管理
package baidu

import (
	"fmt"
	"net/url"

	"aibuddy/pkg/config"
)

// Memory 记忆管理API
type Memory struct {
	client *Client
}

// NewMemory 创建记忆管理客户端
func NewMemory() *Memory {
	return &Memory{client: NewClient()}
}

// ClearCharacterPortraitRequest 清空人物画像请求
type ClearCharacterPortraitRequest struct {
	AppID  string // 互动应用ID
	UserID string // 业务侧用户唯一ID
}

// ClearCharacterPortrait 清空设备人物画像
// 清空当前设备的人物画像信息
func (m *Memory) ClearCharacterPortrait(req *ClearCharacterPortraitRequest) error {
	appID := req.AppID
	if appID == "" {
		appID = config.Instance.Baidu.AppID
	}

	path := "/api/v1/character-portraits"
	query := url.Values{}
	query.Set("appId", appID)
	query.Set("userId", req.UserID)

	if err := m.client.Request("DELETE", path, query, nil, nil); err != nil {
		return fmt.Errorf("清空人物画像失败: %w", err)
	}

	return nil
}
