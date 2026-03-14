package baidu

import (
	"fmt"
	"log/slog"
	"net/url"

	"aibuddy/pkg/config"
)

// TTSVoice TTS复刻音色API
type TTSVoice struct {
	client *Client
}

// NewTTSVoice 创建TTS复刻音色客户端
func NewTTSVoice() *TTSVoice {
	return &TTSVoice{client: NewClient()}
}

// CloneAudio 复刻音频
type CloneAudio struct {
	AudioBytes  string `json:"audio_bytes"`  // 二进制音频字节，需对二进制音频进行base64编码
	AudioFormat string `json:"audio_format"` // wav、mp3
	Text        string `json:"text"`         // 音频文件内容
}

// CreateCloneVoiceRequest 创建复刻音色请求
type CreateCloneVoiceRequest struct {
	AppID        string       `json:"app_id"`                  // 大模型互动应用ID
	UniqID       string       `json:"uniq_id"`                 // C端用户唯一标识
	Name         string       `json:"name,omitempty"`          // 音色名称
	Description  string       `json:"description,omitempty"`   // 音色描述
	AuditionText string       `json:"audition_text,omitempty"` // 试听文本内容
	Audios       []CloneAudio `json:"audios"`                  // 音频列表
	Language     int          `json:"language,omitempty"`      // cn = 0 中文（默认），en = 1 英文，ja = 2 日语
}

// CreateCloneVoiceResponse 创建复刻音色响应
type CreateCloneVoiceResponse struct {
	VoiceID int64 `json:"voice_id"` // 发音人ID
}

// CreateCloneVoice 创建复刻音色
func (t *TTSVoice) CreateCloneVoice(req *CreateCloneVoiceRequest) (*CreateCloneVoiceResponse, error) {
	if req.AppID == "" {
		req.AppID = config.Instance.Baidu.AppID
	}

	path := "/api/v1/ttsVoice/client_clone/create"
	var result CreateCloneVoiceResponse

	if err := t.client.Request("POST", path, nil, req, &result); err != nil {
		slog.Error("[TTSVoice] CreateCloneVoice failed", "err", err)
		return nil, err
	}

	return &result, nil
}

// RetrainCloneVoiceRequest 重新训练复刻音色请求
type RetrainCloneVoiceRequest struct {
	AppID        string       `json:"app_id"`                  // 大模型互动应用ID
	UniqID       string       `json:"uniq_id"`                 // C端用户唯一标识
	Name         string       `json:"name,omitempty"`          // 音色名称
	Description  string       `json:"description,omitempty"`   // 音色描述
	AuditionText string       `json:"audition_text,omitempty"` // 试听文本内容
	Audios       []CloneAudio `json:"audios"`                  // 音频列表
	Language     int          `json:"language,omitempty"`      // cn = 0 中文（默认），en = 1 英文，ja = 2 日语
}

// RetrainCloneVoice 重新训练复刻音色
func (t *TTSVoice) RetrainCloneVoice(voiceID string, req *RetrainCloneVoiceRequest) error {
	if req.AppID == "" {
		req.AppID = config.Instance.Baidu.AppID
	}

	path := fmt.Sprintf("/api/v1/ttsVoice/client_clone/update/%s", voiceID)
	return t.client.Request("PUT", path, nil, req, nil)
}

// CloneVoiceListRequest 获取音色列表请求
type CloneVoiceListRequest struct {
	AppID  string // 大模型互动应用ID
	UniqID string // C端用户唯一标识
}

// CloneVoiceItem 音色项
type CloneVoiceItem struct {
	UniqID      string `json:"uniq_id"`      // 用户唯一标识
	VoiceID     int64  `json:"voice_id"`     // 音色ID
	Name        string `json:"name"`         // 音色名称
	CreateTime  string `json:"create_time"`  // 创建时间
	UpdateTime  string `json:"update_time"`  // 修改时间
	AuditionURL string `json:"audition_url"` // 试听临时文件地址
	Status      string `json:"status"`       // 复刻状态
	Language    string `json:"language"`     // 语言
}

// CloneVoiceListResponse 获取音色列表响应
type CloneVoiceListResponse struct {
	TotalCount int              `json:"total_count"` // 总数量
	Data       []CloneVoiceItem `json:"data"`        // 数据列表
}

// GetCloneVoiceList 获取应用/C端用户下音色列表
func (t *TTSVoice) GetCloneVoiceList(req *CloneVoiceListRequest) (*CloneVoiceListResponse, error) {
	appID := req.AppID
	if appID == "" {
		appID = config.Instance.Baidu.AppID
	}

	path := "/api/v1/ttsVoice/client_clone/list"
	query := url.Values{}
	query.Set("app_id", appID)
	if req.UniqID != "" {
		query.Set("uniq_id", req.UniqID)
	}

	var result CloneVoiceListResponse
	if err := t.client.Request("GET", path, query, nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteCloneVoice 删除C端用户下音色
func (t *TTSVoice) DeleteCloneVoice(appID, uniqID, voiceID string) error {
	if appID == "" {
		appID = config.Instance.Baidu.AppID
	}

	path := fmt.Sprintf("/api/v1/ttsVoice/client_clone/%s/%s/%s", appID, uniqID, voiceID)
	return t.client.Request("DELETE", path, nil, nil, nil)
}
