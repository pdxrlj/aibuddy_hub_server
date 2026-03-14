// Package ttsvoice 提供TTS语音生成功能
package ttsvoice

import (
	"aibuddy/pkg/ahttp"
	"aibuddy/pkg/baidu"
	"aibuddy/pkg/config"
	"net/http"
	"strconv"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = func() trace.Tracer {
	return otel.Tracer(config.Instance.Tracer.ServiceName)
}

// Handler TTS语音处理器
type Handler struct {
	ttsVoice *baidu.TTSVoice
}

// New 创建TTS语音处理器实例
func New() *Handler {
	return &Handler{
		ttsVoice: baidu.NewTTSVoice(),
	}
}

// CreateCloneVoice 创建复刻音色
func (h *Handler) CreateCloneVoice(state *ahttp.State, req *CreateCloneVoiceRequest) error {
	_, span := tracer().Start(state.Ctx.Request().Context(), "tts_voice_create_clone")
	defer span.End()
	span.SetAttributes(attribute.String("uniq_id", req.UniqID))
	span.SetAttributes(attribute.String("name", req.Name))

	audios := make([]baidu.CloneAudio, len(req.Audios))
	for i, audio := range req.Audios {
		audios[i] = baidu.CloneAudio{
			AudioBytes:  audio.AudioBytes,
			AudioFormat: audio.AudioFormat,
			Text:        audio.Text,
		}
	}

	result, err := h.ttsVoice.CreateCloneVoice(&baidu.CreateCloneVoiceRequest{
		AppID:        req.AppID,
		UniqID:       req.UniqID,
		Name:         req.Name,
		Description:  req.Description,
		AuditionText: req.AuditionText,
		Audios:       audios,
		Language:     req.Language,
	})
	if err != nil {
		span.RecordError(err)
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	return state.Resposne().SetData(&CreateCloneVoiceResponse{
		VoiceID: result.VoiceID,
	}).Success()
}

// RetrainCloneVoice 重新训练复刻音色
func (h *Handler) RetrainCloneVoice(state *ahttp.State, req *RetrainCloneVoiceRequest) error {
	_, span := tracer().Start(state.Ctx.Request().Context(), "tts_voice_retrain_clone")
	defer span.End()
	span.SetAttributes(attribute.String("voice_id", req.VoiceID))
	span.SetAttributes(attribute.String("uniq_id", req.UniqID))

	audios := make([]baidu.CloneAudio, len(req.Audios))
	for i, audio := range req.Audios {
		audios[i] = baidu.CloneAudio{
			AudioBytes:  audio.AudioBytes,
			AudioFormat: audio.AudioFormat,
			Text:        audio.Text,
		}
	}

	err := h.ttsVoice.RetrainCloneVoice(req.VoiceID, &baidu.RetrainCloneVoiceRequest{
		AppID:        req.AppID,
		UniqID:       req.UniqID,
		Name:         req.Name,
		Description:  req.Description,
		AuditionText: req.AuditionText,
		Audios:       audios,
		Language:     req.Language,
	})
	if err != nil {
		span.RecordError(err)
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	return state.Resposne().Success()
}

// GetCloneVoiceList 获取音色列表
func (h *Handler) GetCloneVoiceList(state *ahttp.State, req *GetCloneVoiceListRequest) error {
	_, span := tracer().Start(state.Ctx.Request().Context(), "tts_voice_get_list")
	defer span.End()
	span.SetAttributes(attribute.String("uniq_id", req.UniqID))

	result, err := h.ttsVoice.GetCloneVoiceList(&baidu.CloneVoiceListRequest{
		AppID:  req.AppID,
		UniqID: req.UniqID,
	})
	if err != nil {
		span.RecordError(err)
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	items := make([]CloneVoiceItem, len(result.Data))
	for i, item := range result.Data {
		items[i] = CloneVoiceItem{
			UniqID:      item.UniqID,
			VoiceID:     item.VoiceID,
			Name:        item.Name,
			CreateTime:  item.CreateTime,
			UpdateTime:  item.UpdateTime,
			AuditionURL: item.AuditionURL,
			Status:      item.Status,
			Language:    item.Language,
		}
	}

	return state.Resposne().SetData(&GetCloneVoiceListResponse{
		TotalCount: result.TotalCount,
		Data:       items,
	}).Success()
}

// DeleteCloneVoice 删除音色
func (h *Handler) DeleteCloneVoice(state *ahttp.State, req *DeleteCloneVoiceRequest) error {
	_, span := tracer().Start(state.Ctx.Request().Context(), "tts_voice_delete")
	defer span.End()
	span.SetAttributes(attribute.String("uniq_id", req.UniqID))
	span.SetAttributes(attribute.Int64("voice_id", req.VoiceID))

	err := h.ttsVoice.DeleteCloneVoice(req.AppID, req.UniqID, strconv.FormatInt(req.VoiceID, 10))
	if err != nil {
		span.RecordError(err)
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	return state.Resposne().Success()
}
