// Package devicehandler provides the device handler for the server.
package devicehandler

// ================== 公共结构体 ==================

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
	Cloud3AURL            string  `json:"cloud_3a_url,omitempty"`           // 云端3A功能配置
	TTSEnableFastSend     bool    `json:"tts_enable_fast_send,omitempty"`   // TTS/云播放音频加速下发
	TTSFastSendSecond     int     `json:"tts_fast_send_second,omitempty"`   // TTS/云播放音频加速下发持续时长
	TTSFastSendRatio      float64 `json:"tts_fast_send_ratio,omitempty"`    // TTS/云播放音频加速发送倍率
	MCP                   string  `json:"mcp,omitempty"`                    // MCP配置
}

// InstanceContext 实例上下文信息
type InstanceContext struct {
	CID   uint64 `json:"cid"`   // 大模型互动实例sdk使用数据通信ID
	Token string `json:"token"` // 大模型互动实例sdk使用数据通信token
}

// ================== generateAIAgentCall 接口 ==================

// GenerateAIAgentCallRequest 前端创建Agent的请求参数
type GenerateAIAgentCallRequest struct {
	DeviceID     string `param:"device_id" json:"device_id"`
	Config       string `json:"config"`        // 大模型互动实例级别配置，格式是json对象序列化后的string类型
	QuickStart   bool   `json:"quick_start"`   // 快速启动
	AppID        string `json:"app_id"`        // 大模型互动应用ID
	InstanceType string `json:"instance_type"` // 实例类型
}

// GenerateAIAgentCallResponse 创建大模型互动实例响应
type GenerateAIAgentCallResponse struct {
	Code              int              `json:"code"`
	AiAgentInstanceID uint64           `json:"ai_agent_instance_id"`    // AI智能体实例ID
	InstanceType      string           `json:"instance_type,omitempty"` // 实例类型
	ForwardPort       int              `json:"forward_port"`
	Context           *InstanceContext `json:"context,omitempty"` // 实例上下文
}

// ================== stopAIAgentInstance 接口 ==================

// StopAIAgentInstanceRequest 前端销毁Agent请求参数
type StopAIAgentInstanceRequest struct {
	AppID             string `json:"app_id"`               // 应用ID
	AiAgentInstanceID uint64 `json:"ai_agent_instance_id"` // AI智能体实例ID
}

// ================== /userserver/instance/generate 接口 ==================

// AiAgentGenerateRequest AI智能体生成请求参数
type AiAgentGenerateRequest struct {
	AK           string `json:"ak"`            // 鉴权ak，必传
	SK           string `json:"sk"`            // 鉴权sk，必传
	Config       string `json:"config"`        // 大模型互动实例级别配置，格式是json对象序列化后的string类型
	AppID        string `json:"app_id"`        // 大模型互动应用ID
	InstanceType string `json:"instance_type"` // 实例类型
}

// AiAgentGenerateResponse AI智能体生成响应
type AiAgentGenerateResponse struct {
	InstanceID   uint64           `json:"instance_id"`       // 实例ID
	InstanceType string           `json:"instance_type"`     // 实例类型
	Context      *InstanceContext `json:"context,omitempty"` // 实例上下文
}

// ================== /userserver/instance/stop 接口 ==================

// AiAgentDestroyRequest AI智能体销毁请求参数
type AiAgentDestroyRequest struct {
	AppID             string `json:"app_id"`               // 应用ID
	AiAgentInstanceID uint64 `json:"ai_agent_instance_id"` // AI智能体实例ID
	AK                string `json:"ak"`                   // 鉴权ak
	SK                string `json:"sk"`                   // 鉴权sk
}

// ================== /userserver/auth/generate 接口 ==================

// AuthGenerateRequest 鉴权请求参数
type AuthGenerateRequest struct {
	AK              string `json:"ak"`                // 鉴权ak
	SK              string `json:"sk"`                // 鉴权sk
	URL             string `json:"url"`               // 请求URL
	ExpireInSeconds int    `json:"expire_in_seconds"` // 过期时间(秒)
}

// AuthGenerateResponse 鉴权响应
type AuthGenerateResponse struct {
	Authorization string `json:"authorization"` // 鉴权Token
}

// ================== /userserver/instance/baidu、/qianwen、/volc 接口 ==================

// RtcGenerateRequest RTC生成请求参数
type RtcGenerateRequest struct {
	AK             string `json:"ak"`                // 鉴权ak
	SK             string `json:"sk"`                // 鉴权sk
	Config         string `json:"config"`            // 配置
	AppID          string `json:"app_id"`            // 应用ID
	Model          string `json:"model"`             // 模型名称
	WithWebDemoURL bool   `json:"with_web_demo_url"` // 是否返回Web Demo URL
}

// RtcGenerateResponse RTC生成响应
type RtcGenerateResponse struct {
	AiAgentInstanceID uint64           `json:"ai_agent_instance_id"` // AI智能体实例ID
	InstanceType      string           `json:"instance_type"`        // 实例类型
	Context           *InstanceContext `json:"context,omitempty"`    // 实例上下文
	TestURL           string           `json:"test_url,omitempty"`   // 测试URL
}

// ================== switchSceneRole 接口 ==================

// AgentSwitchConfigRequest 切换角色请求参数
type AgentSwitchConfigRequest struct {
	AppID             string `json:"app_id"`               // 应用ID
	AiAgentInstanceID uint64 `json:"ai_agent_instance_id"` // AI智能体实例ID
	SceneRole         string `json:"scene_role"`           // 场景角色
	TTS               string `json:"tts"`                  // TTS配置
	Query             string `json:"query"`                // 查询内容
	TTSSayHi          string `json:"tts_sayhi"`            // 切换角色成功后的打招呼语
}
