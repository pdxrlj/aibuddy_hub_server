// Package baidu 百度云API客户端大模型互动实例管理
package baidu

import (
	"encoding/json"
	"fmt"

	"aibuddy/pkg/config"
	"aibuddy/pkg/helpers"
)

// AIAgent 大模型互动实例API
type AIAgent struct {
	client *Client
}

// NewAIAgent 创建大模型互动实例客户端
func NewAIAgent() *AIAgent {
	return &AIAgent{client: NewClient()}
}

// NewAIAgentWithAKSK 使用指定AK/SK创建大模型互动实例客户端
func NewAIAgentWithAKSK(ak, sk string) *AIAgent {
	return &AIAgent{client: NewClientWithAKSK(ak, sk)}
}

// InstanceType 互动实例类型
type InstanceType string

const (
	// InstanceTypeVoiceChat 语音互动
	InstanceTypeVoiceChat InstanceType = "VoiceChat"
	// InstanceTypeDigitalHuman 数字人互动
	InstanceTypeDigitalHuman InstanceType = "DigitalHuman"
)

// AIAgentConfig 大模型互动实例配置
type AIAgentConfig struct {
	Location              string  `json:"location,omitempty"`               // 用户位置信息
	Role                  string  `json:"role,omitempty"`                   // 简单角色
	SceneRole             string  `json:"sceneRole,omitempty"`              // 复杂场景人设名称
	TTSSayHi              string  `json:"tts_sayhi,omitempty"`              // 招呼语
	Lang                  string  `json:"lang,omitempty"`                   // 语言设置
	AudioCodec            string  `json:"audiocodec,omitempty"`             // 音频格式
	AudioBitrate          int     `json:"audio_bitrate,omitempty"`          // 输出码率
	DisableVoiceAutoInt   bool    `json:"disable_voice_auto_int,omitempty"` // 关闭自动打断
	InterruptionWords     string  `json:"interruption_words,omitempty"`     // 在线打断词
	InterruptionAnswer    string  `json:"interruption_answer,omitempty"`    // 在线打断回复词
	AutoPauseSecs         int     `json:"auto_pause_secs,omitempty"`        // 自动暂停时长
	TTS                   string  `json:"tts,omitempty"`                    // TTS配置名称
	TTSURL                string  `json:"tts_url,omitempty"`                // TTS配置详情
	TTSEndDelayMs         int     `json:"tts_end_delay_ms,omitempty"`       // TTS播报结束延迟时间
	LLM                   string  `json:"llm,omitempty"`                    // 大模型配置
	LLMURL                string  `json:"llm_url,omitempty"`                // 大模型配置URL
	LLMCfg                string  `json:"llm_cfg,omitempty"`                // 大模型配置
	LLMToken              string  `json:"llm_token,omitempty"`              // 大模型Token
	E2ELLMode             string  `json:"e2ellm_mode,omitempty"`            // 语音端到端模型配置
	E2ELLSampleRate       int     `json:"e2ellm_sample_rate,omitempty"`     // 端到端语音模型返回音频采样率
	UserID                string  `json:"user_id,omitempty"`                // 业务侧用户唯一ID
	ContentUsedDB         string  `json:"content_used_db,omitempty"`        // 内容资源私有库
	ASRLongAudioMode      bool    `json:"asr_long_audio_mode,omitempty"`    // ASR按键模式
	ASRVADAppend          bool    `json:"asr_vad_append,omitempty"`         // ASR断句后容许追加
	ASRVAD                int     `json:"asr_vad,omitempty"`                // ASR断句时长(毫秒)
	ASRVADLevel           int     `json:"asr_vad_level,omitempty"`          // ASR说话过程中说话人音量检测
	VoiceFPURL            string  `json:"voice_fp_url,omitempty"`           // 声纹检测配置
	RemoteMusicPlayer     bool    `json:"remote_music_player,omitempty"`    // 音乐播放是否支持云端播放
	EmotionRecognitionCfg string  `json:"emotionRecognitionCfg,omitempty"`  // 情绪开关配置
	Cloud3AURL            string  `json:"cloud_3A_url,omitempty"`           // 云端3A功能配置
	TTSEnableFastSend     bool    `json:"tts_enable_fast_send,omitempty"`   // TTS/云播放音频加速下发
	TTSFastSendSecond     int     `json:"tts_fast_send_second,omitempty"`   // TTS/云播放音频加速下发持续时长
	TTSFastSendRatio      float64 `json:"tts_fast_send_ratio,omitempty"`    // TTS/云播放音频加速发送倍率
	MCP                   string  `json:"mcp,omitempty"`                    // MCP配置
}

// GenerateAIAgentCallRequest 创建大模型互动实例请求
type GenerateAIAgentCallRequest struct {
	AppID        string       `json:"app_id"`
	InstanceType InstanceType `json:"instance_type,omitempty"`
	Config       any          `json:"config,omitempty"` // 支持 *AIAgentConfig 或 string
}

// AIAgentContext 大模型互动实例上下文
type AIAgentContext struct {
	CID   uint64 `json:"cid"`
	Token string `json:"token"`
}

// GenerateAIAgentCallResponse 创建大模型互动实例响应
type GenerateAIAgentCallResponse struct {
	AiAgentInstanceID uint64          `json:"ai_agent_instance_id"`
	InstanceType      string          `json:"instance_type"`
	Context           *AIAgentContext `json:"context,omitempty"`
	XBCERequestID     string          `json:"-"`
}

// GenerateAIAgentCall 创建并启动大模型互动实例
func (a *AIAgent) GenerateAIAgentCall(req *GenerateAIAgentCallRequest) (*GenerateAIAgentCallResponse, error) {
	appID := req.AppID
	if appID == "" {
		appID = config.Instance.Baidu.AppID
	}

	path := "/api/v1/aiagent/generateAIAgentCall"
	body := map[string]any{
		"app_id": appID,
	}

	if req.InstanceType != "" {
		body["instance_type"] = req.InstanceType
	}

	if req.Config != nil {
		switch cfg := req.Config.(type) {
		case string:
			// 如果已经是字符串，直接传递（原始JSON字符串）
			body["config"] = cfg
		case *AIAgentConfig:
			// 如果是结构体，需要序列化
			configBytes, err := json.Marshal(cfg)
			if err != nil {
				return nil, fmt.Errorf("序列化配置失败: %w", err)
			}
			body["config"] = string(configBytes)
		}
	}
	fmt.Println("=========AIAgent=========")
	helpers.PP(body)
	var result GenerateAIAgentCallResponse
	requestID, err := a.client.RequestWithHeader("POST", path, nil, body, &result, "x-bce-request-id")
	if err != nil {
		return nil, fmt.Errorf("创建大模型互动实例失败: %w", err)
	}
	result.XBCERequestID = requestID

	return &result, nil
}

// StopAIAgentInstanceRequest 停止大模型互动实例请求
type StopAIAgentInstanceRequest struct {
	AppID             string
	AiAgentInstanceID any `json:"ai_agent_instance_id"` // 大模型互动实例ID，支持string或uint64
}

// StopAIAgentInstance 停止大模型互动实例
func (a *AIAgent) StopAIAgentInstance(req *StopAIAgentInstanceRequest) error {
	appID := req.AppID
	if appID == "" {
		appID = config.Instance.Baidu.AppID
	}

	path := "/api/v1/aiagent/stopAIAgentInstance"
	body := map[string]any{
		"app_id":               appID,
		"ai_agent_instance_id": req.AiAgentInstanceID,
	}

	if err := a.client.Request("POST", path, nil, body, nil); err != nil {
		return fmt.Errorf("停止大模型互动实例失败: %w", err)
	}

	return nil
}

// InterruptRequest 打断大模型互动实例播报请求
type InterruptRequest struct {
	AppID             string
	AiAgentInstanceID uint64 `json:"ai_agent_instance_id"`
	ExtraMsg          string `json:"extra_msg,omitempty"`
}

// Interrupt 服务端打断大模型互动实例播报
func (a *AIAgent) Interrupt(req *InterruptRequest) error {
	appID := req.AppID
	if appID == "" {
		appID = config.Instance.Baidu.AppID
	}

	path := "/api/v1/aiagent/interrupt"
	body := map[string]any{
		"app_id":               appID,
		"ai_agent_instance_id": req.AiAgentInstanceID,
	}

	if req.ExtraMsg != "" {
		body["extra_msg"] = req.ExtraMsg
	}

	if err := a.client.Request("POST", path, nil, body, nil); err != nil {
		return fmt.Errorf("打断大模型互动实例播报失败: %w", err)
	}

	return nil
}

// SendMsgRequest 大模型互动实例发消息给SDK请求
type SendMsgRequest struct {
	AppID             string
	AiAgentInstanceID uint64 `json:"ai_agent_instance_id"`
	Message           string `json:"message"`
}

// SendMsg 大模型互动实例发消息给SDK
func (a *AIAgent) SendMsg(req *SendMsgRequest) error {
	appID := req.AppID
	if appID == "" {
		appID = config.Instance.Baidu.AppID
	}

	path := "/api/v1/aiagent/sendMsg"
	body := map[string]any{
		"app_id":               appID,
		"ai_agent_instance_id": req.AiAgentInstanceID,
		"message":              req.Message,
	}

	if err := a.client.Request("POST", path, nil, body, nil); err != nil {
		return fmt.Errorf("发送消息给SDK失败: %w", err)
	}

	return nil
}
