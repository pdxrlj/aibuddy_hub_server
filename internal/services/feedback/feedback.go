// Package feedback defines the services for the feedback.
package feedback

import (
	"aibuddy/internal/model"
	"aibuddy/internal/repository"
	"context"
)

// Service 反馈服务
type Service struct {
	FeedbackRepository *repository.FeedbackRepository
}

// NewService 创建反馈服务
func NewService() *Service {
	return &Service{FeedbackRepository: repository.NewFeedbackRepository()}
}

// Create 创建反馈
func (s *Service) Create(ctx context.Context, feedback *model.Feedback) error {
	return s.FeedbackRepository.Create(ctx, feedback)
}
