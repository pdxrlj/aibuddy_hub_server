// Package baidu 百度云服务层
package baidu

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"aibuddy/internal/repository"
	"aibuddy/internal/services/cache"
	"aibuddy/internal/services/role"
	"aibuddy/pkg/baidu"
	"aibuddy/pkg/config"

	"log/slog"

	"github.com/spf13/cast"
)

// ConfigRequest 配置请求参数
type ConfigRequest struct {
	LLM                 string `json:"llm,omitempty"`
	LLMToken            string `json:"llm_token,omitempty"`
	TTSURL              string `json:"tts_url,omitempty"`
	TTSPrivateServerURL string `json:"tts_private_server_url,omitempty"`

	RTCAC             string `json:"rtc_ac,omitempty"`
	Lang              string `json:"lang,omitempty"`
	RemoteMusicPlayer bool   `json:"remote_music_player,omitempty"`
	EnableVisual      string `json:"enable_visual,omitempty"`
	DFDA              string `json:"dfda,omitempty"`
	TTS               string `json:"tts,omitempty"`
	TTSEndDelayMs     int    `json:"tts_end_delay_ms,omitempty"`

	EmotionRecognitionCfg *EmotionRecognitionCfg `json:"emotion_recognition_cfg,omitempty"`
}

// EmotionRecognitionCfg 情感识别配置
type EmotionRecognitionCfg struct {
	Enable         bool     `json:"enable,omitempty"`
	InjectToLLM    bool     `json:"inject_to_llm,omitempty"`
	TTSWithEmotion bool     `json:"tts_with_emotion,omitempty"`
	EmotionList    []string `json:"emotion_list,omitempty"`
}

// GenerateAIAgentCallRequest 创建AIAgent请求
type GenerateAIAgentCallRequest struct {
	AppID        string
	InstanceType string
	Config       *ConfigRequest
	DeviceID     string
}

// GenerateAIAgentCallResponse 创建AIAgent响应
type GenerateAIAgentCallResponse struct {
	AiAgentInstanceID uint64
	InstanceType      string
	Context           *baidu.AIAgentContext
}

// Service 百度云服务
type Service struct {
	aiAgent     *baidu.AIAgent
	RoleService *role.Service

	DeviceRepo *repository.DeviceRepo
}

// NewService 创建百度云服务实例
func NewService() *Service {
	return &Service{
		aiAgent:     baidu.NewAIAgent(),
		RoleService: role.NewRoleService(),
		DeviceRepo:  repository.NewDeviceRepo(),
	}
}

// getDefaultAppID 获取默认的AppID
func getDefaultAppID(appID string) string {
	if appID == "" {
		return config.Instance.Baidu.AppID
	}
	return appID
}

// buildConfigStr 构建配置字符串
func buildConfigStr(cfg *ConfigRequest, voiceID string) (string, error) {
	if cfg == nil {
		return "", nil
	}

	if cfg.TTS == "" {
		// cfg.TTS = "PRIVATE_EXTEND" // DEFAULT 默认音色
		cfg.TTS = "DEFAULT"
	}

	vcn := config.Instance.Baidu.TTS.Vcn
	if voiceID != "" {
		vcn = voiceID
	}

	if cfg.TTSURL == "" {
		// cfg.TTSURL = fmt.Sprintf(`DEFAULT{"vcn":"%s","vol":2.0,"spd":1.0,"emotion":"fluent"}`, vcn)
		// cfg.TTSURL = fmt.Sprintf(`DEFAULT{"vcn":"%s","emotion":"calm","pit":-1}`, vcn)
		cfg.TTSURL = fmt.Sprintf(`PRIVATE_EXTEND{"vcn":"%s"}`, vcn)
	}

	if cfg.TTSPrivateServerURL == "" {
		cfg.TTSPrivateServerURL = "ws://8.153.82.116:8289/ws/2.0/speech/publiccloudspeech/v1/tts"
	}

	if cfg.TTSEndDelayMs == 0 {
		TTsEndDelayMs := config.Instance.Baidu.TTS.TtsEndDelayMs
		if TTsEndDelayMs == 0 {
			TTsEndDelayMs = 50
		}
		cfg.TTSEndDelayMs = TTsEndDelayMs
	}

	if cfg.EmotionRecognitionCfg == nil {
		cfg.EmotionRecognitionCfg = &EmotionRecognitionCfg{
			Enable:         true,
			InjectToLLM:    true,
			TTSWithEmotion: true,
		}
	}

	configBytes, err := json.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("序列化配置失败: %w", err)
	}
	return string(configBytes), nil
}

// GenerateAIAgentCall 创建并启动大模型互动实例
func (s *Service) GenerateAIAgentCall(ctx context.Context, req *GenerateAIAgentCallRequest) (*GenerateAIAgentCallResponse, error) {
	expired, err := s.DeviceRepo.IsDeviceVipExpired(ctx, req.DeviceID)
	if err != nil {
		return nil, err
	}

	if expired {
		return nil, errors.New("当前设备已经过期了，请充值后在使用")
	}

	device, err := s.DeviceRepo.GetDeviceInfo(ctx, req.DeviceID)
	if err != nil {
		return nil, err
	}

	appID := getDefaultAppID(req.AppID)

	configStr, err := buildConfigStr(req.Config, device.VoiceID)
	if err != nil {
		return nil, err
	}

	request := &baidu.GenerateAIAgentCallRequest{
		AppID:        appID,
		InstanceType: baidu.InstanceType(req.InstanceType),
		Config:       configStr,
	}

	slog.Info("[RTC] GenerateAIAgentCall", "request", request)

	resp, err := s.aiAgent.GenerateAIAgentCall(request)
	if err != nil {
		slog.Error("Failed to create AIAgentInstance", "error", err)
		return nil, err
	}

	// 保存当前会话RTC实例ID
	slog.Info("[RTC] Store RTC instance ID", "deviceID", req.DeviceID, "instanceID", resp.AiAgentInstanceID)
	_ = cache.StoreRTCInstanceID(req.DeviceID, cast.ToString(resp.AiAgentInstanceID))

	// 切换到小程序选择的角色
	err = s.RoleService.DeviceInstanceSwitchDefRole(ctx, resp.AiAgentInstanceID, req.DeviceID)
	if err != nil {
		slog.Error("Failed to switch default role", "error", err)
		return nil, fmt.Errorf("切换默认角色失败: %w", err)
	}

	return &GenerateAIAgentCallResponse{
		AiAgentInstanceID: resp.AiAgentInstanceID,
		InstanceType:      resp.InstanceType,
		Context:           resp.Context,
	}, nil
}
