// Package emotion is the handler for the emotion trigger
package emotion

import (
	"aibuddy/pkg/ahttp"
	"aibuddy/pkg/config"
	"errors"

	emotion "aibuddy/internal/services/emotion"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

var tracer = func() trace.Tracer {
	return otel.Tracer(config.Instance.Tracer.ServiceName)
}

// Handler is the handler for the emotion trigger
type Handler struct {
	EmotionService *emotion.Service
}

// NewHandler creates a new emotion trigger handler
func NewHandler() *Handler {
	return &Handler{
		EmotionService: emotion.NewService(),
	}
}

// GetEmotions gets the emotion triggers
func (h *Handler) GetEmotions(state *ahttp.State, req *GetEmotionsRequest) error {
	ctx, span := tracer().Start(state.Context(), "EmotionTriggerHandler.GetEmotions")
	defer span.End()

	emotions, total, err := h.EmotionService.GetEmotions(ctx, req.Page, req.PageSize, req.DeviceID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", req.DeviceID))
		span.SetAttributes(attribute.Int("page", req.Page))
		span.SetAttributes(attribute.Int("page_size", req.PageSize))
		return state.Resposne().Error(err)
	}
	return state.Resposne().Success(GetEmotionsRes{
		Data:  emotions,
		Total: total,
	})
}

// GetLatestEmotion gets the latest emotion trigger
func (h *Handler) GetLatestEmotion(state *ahttp.State, req *GetLatestEmotionRequest) error {
	ctx, span := tracer().Start(state.Context(), "EmotionTriggerHandler.GetLatestEmotion")
	defer span.End()

	emotion, err := h.EmotionService.GetLatestEmotion(ctx, req.DeviceID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", req.DeviceID))
		return state.Resposne().Error(err)
	}
	return state.Resposne().Success(GetLatestEmotionRes{
		Data: emotion,
	})
}
