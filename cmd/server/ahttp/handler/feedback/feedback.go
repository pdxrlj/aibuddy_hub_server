// Package feedback defines the HTTP handlers for the feedback.
package feedback

import (
	"aibuddy/internal/model"
	aiuserService "aibuddy/internal/services/aiuser"
	"aibuddy/internal/services/feedback"
	"aibuddy/pkg/ahttp"
	"aibuddy/pkg/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = func() trace.Tracer {
	return otel.Tracer(config.Instance.Tracer.ServiceName)
}

// Handler 反馈处理器
type Handler struct {
	feedbackService *feedback.Service
}

// NewHandler 创建反馈处理器
func NewHandler() *Handler {
	return &Handler{
		feedbackService: feedback.NewService(),
	}
}

// Create 创建反馈
func (h *Handler) Create(state *ahttp.State, req *CreateFeedbackRequest) error {
	ctx, span := tracer().Start(state.Context(), "CreateFeedback")
	defer span.End()

	uid, err := aiuserService.GetUIDFromContext(state.Ctx)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("feedback_type", req.FeedbackType), attribute.String("content", req.Content))
		return state.Response().Error(err)
	}

	images, err := req.ImagesEncode()
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.Int64("uid", uid), attribute.String("feedback_type", req.FeedbackType), attribute.String("content", req.Content))
		return state.Response().Error(err)
	}

	feedback := &model.Feedback{
		FeedbackType: model.FeedbackType(req.FeedbackType),
		Content:      req.Content,
		UID:          uid,
		Images:       images,
	}
	err = h.feedbackService.Create(ctx, feedback)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(
			attribute.Int64("uid", uid),
			attribute.String("feedback_type", req.FeedbackType),
			attribute.String("content", req.Content),
			attribute.String("images", string(images)),
		)
		return state.Response().Error(err)
	}
	return state.Response().Success()
}
