// Package ttsvoice TTS语音请求响应结构体
package ttsvoice

// CloneAudio 复刻音频
type CloneAudio struct {
	AudioBytes  string `json:"audio_bytes"`  // 二进制音频字节，需对二进制音频进行base64编码
	AudioFormat string `json:"audio_format"` // wav、mp3
	Text        string `json:"text"`         // 音频文件内容
}

// CreateCloneVoiceRequest 创建复刻音色请求
type CreateCloneVoiceRequest struct {
	AppID        string       `json:"app_id"`                                                // 大模型互动应用ID
	UniqID       string       `json:"uniq_id" validate:"required" msg:"required:用户唯一标识不能为空"` // C端用户唯一标识
	Name         string       `json:"name"`                                                  // 音色名称
	Description  string       `json:"description"`                                           // 音色描述
	AuditionText string       `json:"audition_text"`                                         // 试听文本内容
	Audios       []CloneAudio `json:"audios" validate:"required" msg:"required:音频列表不能为空"`    // 音频列表
	Language     int          `json:"language"`                                              // cn = 0 中文（默认），en = 1 英文，ja = 2 日语
}

// CreateCloneVoiceResponse 创建复刻音色响应
type CreateCloneVoiceResponse struct {
	VoiceID int64 `json:"voice_id"` // 发音人ID
}

// RetrainCloneVoiceRequest 重新训练复刻音色请求
type RetrainCloneVoiceRequest struct {
	VoiceID      string       `param:"voice_id" validate:"required" msg:"required:音色ID不能为空"` // 音色ID
	AppID        string       `json:"app_id"`                                                // 大模型互动应用ID
	UniqID       string       `json:"uniq_id" validate:"required" msg:"required:用户唯一标识不能为空"` // C端用户唯一标识
	Name         string       `json:"name"`                                                  // 音色名称
	Description  string       `json:"description"`                                           // 音色描述
	AuditionText string       `json:"audition_text"`                                         // 试听文本内容
	Audios       []CloneAudio `json:"audios" validate:"required" msg:"required:音频列表不能为空"`    // 音频列表
	Language     int          `json:"language"`                                              // cn = 0 中文（默认），en = 1 英文，ja = 2 日语
}

// GetCloneVoiceListRequest 获取音色列表请求
type GetCloneVoiceListRequest struct {
	AppID  string `query:"app_id"`                                                // 大模型互动应用ID
	UniqID string `query:"uniq_id" validate:"required" msg:"required:用户唯一标识不能为空"` // C端用户唯一标识
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

// GetCloneVoiceListResponse 获取音色列表响应
type GetCloneVoiceListResponse struct {
	TotalCount int              `json:"total_count"` // 总数量
	Data       []CloneVoiceItem `json:"data"`        // 数据列表
}

// DeleteCloneVoiceRequest 删除音色请求
type DeleteCloneVoiceRequest struct {
	UniqID  string `param:"uniq_id" validate:"required" msg:"required:用户唯一标识不能为空"` // C端用户唯一标识
	VoiceID int64  `param:"voice_id" validate:"required" msg:"required:音色ID不能为空"`  // 音色ID
	AppID   string `query:"app_id"`                                                // 大模型互动应用ID
}
