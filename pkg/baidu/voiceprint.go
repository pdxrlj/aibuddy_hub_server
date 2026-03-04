// Package baidu 百度云API客户端声纹管理
package baidu

import (
	"fmt"
	"net/url"

	"aibuddy/pkg/config"
)

// Voiceprint 声纹管理API
type Voiceprint struct {
	client *Client
}

// NewVoiceprint 创建声纹管理客户端
func NewVoiceprint() *Voiceprint {
	return &Voiceprint{client: NewClient()}
}

// VoiceprintFormat 声纹音频格式
type VoiceprintFormat string

const (
	// VoiceprintFormatPCM PCM格式
	VoiceprintFormatPCM VoiceprintFormat = "PCM"
	// VoiceprintFormatWAV WAV格式
	VoiceprintFormatWAV VoiceprintFormat = "WAV"
)

// RegisterVoiceprintRequest 注册声纹请求
type RegisterVoiceprintRequest struct {
	AppID      string           // 大模型互动应用ID（可选，默认使用配置）
	NickName   string           `json:"nick_name"`   // 用户昵称，用于匹配时返回信息
	UserID     string           `json:"user_id"`     // 用户唯一标识
	Format     VoiceprintFormat `json:"format"`      // 音频格式：PCM/WAV，要求单声道16k
	FileBase64 string           `json:"file_base64"` // 文件base64编码
}

// RegisterVoiceprintResponse 注册声纹响应
type RegisterVoiceprintResponse struct {
	AppID    string `json:"app_id"`    // 大模型互动应用ID
	NickName string `json:"nick_name"` // 用户昵称
	UserID   string `json:"user_id"`   // 用户唯一标识
	VpID     string `json:"vp_id"`     // 声纹ID
}

// RegisterVoiceprint 注册声纹
func (v *Voiceprint) RegisterVoiceprint(req *RegisterVoiceprintRequest) (*RegisterVoiceprintResponse, error) {
	appID := req.AppID
	if appID == "" {
		appID = config.Instance.Baidu.AppID
	}

	path := "/api/v1/voiceprint/register"
	body := map[string]any{
		"app_id":      appID,
		"nick_name":   req.NickName,
		"user_id":     req.UserID,
		"format":      req.Format,
		"file_base64": req.FileBase64,
	}

	var result RegisterVoiceprintResponse
	if err := v.client.Request("POST", path, nil, body, &result); err != nil {
		return nil, fmt.Errorf("注册声纹失败: %w", err)
	}

	return &result, nil
}

// DeleteVoiceprintRequest 删除声纹请求
type DeleteVoiceprintRequest struct {
	AppID  string   // 大模型互动应用ID（可选，默认使用配置）
	UserID string   `json:"user_id"` // 用户唯一标识
	VpIDs  []string `json:"vp_ids"`  // 批量删除的声纹ID列表
}

// DeleteVoiceprint 删除声纹
func (v *Voiceprint) DeleteVoiceprint(req *DeleteVoiceprintRequest) error {
	appID := req.AppID
	if appID == "" {
		appID = config.Instance.Baidu.AppID
	}

	path := "/api/v1/voiceprint/delete"
	body := map[string]any{
		"app_id":  appID,
		"user_id": req.UserID,
		"vp_ids":  req.VpIDs,
	}

	if err := v.client.Request("DELETE", path, nil, body, nil); err != nil {
		return fmt.Errorf("删除声纹失败: %w", err)
	}

	return nil
}

// VoiceprintItem 声纹项
type VoiceprintItem struct {
	ID       string `json:"id"`        // 声纹ID
	AppID    string `json:"app_id"`    // 应用ID
	NickName string `json:"nick_name"` // 用户昵称
	UserID   string `json:"user_id"`   // 用户唯一标识
	Format   string `json:"format"`    // 音频格式
	Timestamp int64  `json:"timestamp"` // 时间戳
}

// ListVoiceprintResponse 查询声纹列表响应
type ListVoiceprintResponse struct {
	Data []VoiceprintItem `json:"data"` // 声纹列表
}

// ListVoiceprintRequest 查询声纹列表请求
type ListVoiceprintRequest struct {
	AppID  string // 大模型互动应用ID（可选，默认使用配置）
	UserID string // 用户唯一标识
}

// ListVoiceprint 查询声纹列表
func (v *Voiceprint) ListVoiceprint(req *ListVoiceprintRequest) (*ListVoiceprintResponse, error) {
	appID := req.AppID
	if appID == "" {
		appID = config.Instance.Baidu.AppID
	}

	path := "/api/v1/voiceprint/list"
	query := url.Values{}
	query.Set("app_id", appID)
	query.Set("user_id", req.UserID)

	var result ListVoiceprintResponse
	if err := v.client.Request("GET", path, query, nil, &result); err != nil {
		return nil, fmt.Errorf("查询声纹列表失败: %w", err)
	}

	return &result, nil
}
