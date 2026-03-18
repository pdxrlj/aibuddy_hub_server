// Package devicehandler provides the device handler for the server.
package devicehandler

import (
	"encoding/json"
	"errors"
	"fmt"

	baiduservice "aibuddy/internal/services/baidu"
	"aibuddy/internal/services/cache"
	"aibuddy/internal/services/role"
	"aibuddy/pkg/ahttp"
	"aibuddy/pkg/baidu"
	"aibuddy/pkg/config"
	"aibuddy/pkg/flash"
	"log/slog"
)

const (
	rtcTestURL = "http://brtc-sdk.bj.bcebos.com/web/demo/brtc_chat.html?a=%s&r=%s&u=%s&token=%s"
	location   = "北京市海淀区"
	lang       = "zh"
	volcPrefix = "VOLC"
	volcVcn    = "zh_female_daimengchuanmei_moon_bigtts"
	volcVol    = 1.0
	volcSpd    = 1.0
)

// RtcHandler is the handler for the RTC service.
type RtcHandler struct {
	aiAgent      *baidu.AIAgent
	switchRole   *baidu.SwitchRole
	Cache        flash.Flash
	RoleService  *role.Service
	BaiduService *baiduservice.Service
}

// NewRtcHandler creates a new RtcHandler.
func NewRtcHandler() *RtcHandler {
	return &RtcHandler{
		aiAgent:      baidu.NewAIAgent(),
		switchRole:   baidu.NewSwitchRole(),
		Cache:        cache.Flash(),
		RoleService:  role.NewRoleService(),
		BaiduService: baiduservice.NewService(),
	}
}

// GenerateAIAgentCall 与端侧SDK交互，创建AIAgentInstance
func (h *RtcHandler) GenerateAIAgentCall(state *ahttp.State, req *GenerateAIAgentCallRequest) error {
	if req.CustomSelfCfg == nil || req.CustomSelfCfg.DeviceID == "" {
		return state.Resposne().Error(errors.New("缺少自定义的设备的ID"))
	}

	resp, err := h.BaiduService.GenerateAIAgentCall(state.Ctx.Request().Context(),
		&baiduservice.GenerateAIAgentCallRequest{
			AppID:        req.AppID,
			InstanceType: req.InstanceType,
			Config: &baiduservice.ConfigRequest{
				LLM:               req.Config.LLM,
				LLMToken:          req.Config.LLMToken,
				TTSURL:            req.Config.TTSURL,
				RTCAC:             req.Config.RTCAC,
				Lang:              req.Config.Lang,
				RemoteMusicPlayer: req.Config.RemoteMusicPlayer,
				EnableVisual:      req.Config.EnableVisual,
				DFDA:              req.Config.DFDA,
				TTS:               req.Config.TTS,
				TTSEndDelayMs:     req.Config.TTSEndDelayMS,
			},
			DeviceID: req.CustomSelfCfg.DeviceID,
		})
	if err != nil {
		return state.Resposne().Error(err)
	}

	return state.Resposne().
		Raw(&GenerateAIAgentCallResponse{
			Code:              0,
			AiAgentInstanceID: resp.AiAgentInstanceID,
			ForwardPort:       0,
			Context:           convertInstanceContext(resp.Context),
		})
}

// StopAIAgentInstance 与端侧SDK交互，停止AIAgentInstance
func (h *RtcHandler) StopAIAgentInstance(state *ahttp.State, req *StopAIAgentInstanceRequest) error {
	appID := req.AppID
	if appID == "" {
		appID = config.Instance.Baidu.AppID
	}

	err := h.aiAgent.StopAIAgentInstance(&baidu.StopAIAgentInstanceRequest{
		AppID:             appID,
		AiAgentInstanceID: req.AiAgentInstanceID,
	})
	if err != nil {
		slog.Error("Failed to stop AIAgentInstance", "error", err)
		return state.Resposne().Error(err)
	}

	return state.Resposne().Success()
}

// InstanceGenerate 创建AI智能体互动实例，返回token和实例上下文
func (h *RtcHandler) InstanceGenerate(state *ahttp.State, req *AiAgentGenerateRequest) error {
	aiAgent := baidu.NewAIAgentWithAKSK(req.AK, req.SK)

	resp, err := aiAgent.GenerateAIAgentCall(&baidu.GenerateAIAgentCallRequest{
		AppID:        req.AppID,
		InstanceType: baidu.InstanceType(req.InstanceType),
		Config:       req.Config,
	})
	if err != nil {
		slog.Error("Failed to create AIAgentInstance", "error", err)
		return state.Resposne().Error(err)
	}

	return state.Resposne().Raw(&AiAgentGenerateResponse{
		InstanceID:   resp.AiAgentInstanceID,
		InstanceType: resp.InstanceType,
		Context:      convertInstanceContext(resp.Context),
	})
}

// InstanceStop 销毁AI智能体互动实例
func (h *RtcHandler) InstanceStop(state *ahttp.State, req *AiAgentDestroyRequest) error {
	aiAgent := baidu.NewAIAgentWithAKSK(req.AK, req.SK)

	err := aiAgent.StopAIAgentInstance(&baidu.StopAIAgentInstanceRequest{
		AppID:             req.AppID,
		AiAgentInstanceID: req.AiAgentInstanceID,
	})
	if err != nil {
		slog.Error("Failed to stop AIAgentInstance", "error", err)
		return state.Resposne().Error(err)
	}

	return state.Resposne().Success()
}

// AuthGenerate 获取RTC服务的Token
func (h *RtcHandler) AuthGenerate(state *ahttp.State, req *AuthGenerateRequest) error {
	client := baidu.NewClientWithAKSK(req.AK, req.SK)
	authorization, err := client.BuildAuthorization("POST", req.URL)
	if err != nil {
		slog.Error("Failed to generate Authorization", "error", err)
		return state.Resposne().Error(err)
	}

	return state.Resposne().Raw(&AuthGenerateResponse{
		Authorization: authorization,
	})
}

// InstanceBaidu 使用百度千帆大模型创建RTC实例
func (h *RtcHandler) InstanceBaidu(state *ahttp.State, req *RtcGenerateRequest) error {
	aiAgent := baidu.NewAIAgentWithAKSK(req.AK, req.SK)

	// 构建千帆配置
	cfg := config.Instance.Baidu
	model := req.Model
	if model == "" && cfg.Qianfan != nil {
		model = cfg.Qianfan.Model
	}

	configMap := map[string]string{
		"lang":      lang,
		"location":  location,
		"llm":       "OPENAI",
		"llm_url":   cfg.Qianfan.BaseURL,
		"llm_cfg":   model,
		"llm_token": cfg.Qianfan.APIKey,
	}
	configBytes, _ := json.Marshal(configMap)

	resp, err := aiAgent.GenerateAIAgentCall(&baidu.GenerateAIAgentCallRequest{
		AppID:  req.AppID,
		Config: string(configBytes),
	})
	if err != nil {
		slog.Error("Failed to create AIAgentInstance with Baidu model", "error", err)
		return state.Resposne().Error(err)
	}

	response := &RtcGenerateResponse{
		AiAgentInstanceID: resp.AiAgentInstanceID,
		InstanceType:      resp.InstanceType,
		Context:           convertInstanceContext(resp.Context),
	}

	if req.WithWebDemoURL && resp.Context != nil {
		response.TestURL = formatTestURL(req.AppID, resp.AiAgentInstanceID, resp.Context.Token)
	}

	return state.Resposne().Raw(response)
}

// InstanceQianwen 使用千问大模型创建RTC实例
func (h *RtcHandler) InstanceQianwen(state *ahttp.State, req *RtcGenerateRequest) error {
	aiAgent := baidu.NewAIAgentWithAKSK(req.AK, req.SK)

	cfg := config.Instance.Baidu
	model := req.Model
	if model == "" && cfg.Qianwen != nil {
		model = cfg.Qianwen.Model
	}

	configMap := map[string]string{
		"lang":      lang,
		"llm":       "OPENAI",
		"llm_url":   cfg.Qianwen.BaseURL,
		"llm_cfg":   model,
		"llm_token": cfg.Qianwen.APIKey,
	}
	configBytes, _ := json.Marshal(configMap)

	resp, err := aiAgent.GenerateAIAgentCall(&baidu.GenerateAIAgentCallRequest{
		AppID:  req.AppID,
		Config: string(configBytes),
	})
	if err != nil {
		slog.Error("Failed to create AIAgentInstance with Qianwen model", "error", err)
		return state.Resposne().Error(err)
	}

	response := &RtcGenerateResponse{
		AiAgentInstanceID: resp.AiAgentInstanceID,
		InstanceType:      resp.InstanceType,
		Context:           convertInstanceContext(resp.Context),
	}

	if req.WithWebDemoURL && resp.Context != nil {
		response.TestURL = formatTestURL(req.AppID, resp.AiAgentInstanceID, resp.Context.Token)
	}

	return state.Resposne().Raw(response)
}

// InstanceVolc 使用VOLC TTS创建RTC实例
func (h *RtcHandler) InstanceVolc(state *ahttp.State, req *RtcGenerateRequest) error {
	aiAgent := baidu.NewAIAgentWithAKSK(req.AK, req.SK)

	cfg := config.Instance.Baidu
	vcn := volcVcn
	vol := volcVol
	spd := volcSpd
	if cfg.Volc != nil {
		if cfg.Volc.Vcn != "" {
			vcn = cfg.Volc.Vcn
		}
		if cfg.Volc.Vol > 0 {
			vol = cfg.Volc.Vol
		}
		if cfg.Volc.Spd > 0 {
			spd = cfg.Volc.Spd
		}
	}

	ttsConfigMap := map[string]any{
		"vcn":    vcn,
		"vol":    vol,
		"spd":    spd,
		"apid":   cfg.Volc.Apid,
		"apikey": cfg.Volc.Apikey,
	}
	ttsConfigBytes, _ := json.Marshal(ttsConfigMap)

	configMap := map[string]any{
		"lang":    lang,
		"tts":     volcPrefix,
		"tts_url": volcPrefix + string(ttsConfigBytes),
	}
	configBytes, _ := json.Marshal(configMap)

	resp, err := aiAgent.GenerateAIAgentCall(&baidu.GenerateAIAgentCallRequest{
		AppID:  req.AppID,
		Config: string(configBytes),
	})
	if err != nil {
		slog.Error("Failed to create AIAgentInstance with VOLC TTS", "error", err)
		return state.Resposne().Error(err)
	}

	response := &RtcGenerateResponse{
		AiAgentInstanceID: resp.AiAgentInstanceID,
		InstanceType:      resp.InstanceType,
		Context:           convertInstanceContext(resp.Context),
	}

	if req.WithWebDemoURL && resp.Context != nil {
		response.TestURL = formatTestURL(req.AppID, resp.AiAgentInstanceID, resp.Context.Token)
	}

	return state.Resposne().Raw(response)
}

// SwitchSceneRole 切换角色（音色）
func (h *RtcHandler) SwitchSceneRole(state *ahttp.State, req *AgentSwitchConfigRequest) error {
	appID := req.AppID
	if appID == "" {
		appID = config.Instance.Baidu.AppID
	}

	err := h.switchRole.SwitchSceneRole(&baidu.SwitchRoleRequest{
		AppID:             appID,
		AiAgentInstanceID: req.AiAgentInstanceID,
		SceneRole:         req.SceneRole,
		TTS:               req.TTS,
		Query:             req.Query,
		TTSSayHi:          req.TTSSayHi,
	})
	if err != nil {
		slog.Error("Failed to switch scene role", "error", err)
		return state.Resposne().Error(err)
	}

	return state.Resposne().Success()
}

// convertInstanceContext 转换实例上下文
func convertInstanceContext(ctx *baidu.AIAgentContext) *InstanceContext {
	if ctx == nil {
		return nil
	}
	return &InstanceContext{
		CID:   ctx.CID,
		Token: ctx.Token,
	}
}

// formatTestURL 格式化测试URL
func formatTestURL(appID string, instanceID uint64, token string) string {
	instanceIDStr := fmt.Sprintf("%d", instanceID)
	return fmt.Sprintf(rtcTestURL, appID, instanceIDStr, instanceIDStr, token)
}
