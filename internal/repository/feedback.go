// Package repository defines the data repositories for the application.
package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"context"
)

// FeedbackRepository 反馈仓库
type FeedbackRepository struct{}

// NewFeedbackRepository 创建反馈仓库
func NewFeedbackRepository() *FeedbackRepository {
	return &FeedbackRepository{}
}

// Create 创建反馈
func (r *FeedbackRepository) Create(ctx context.Context, feedback *model.Feedback) error {
	return query.Feedback.WithContext(ctx).Create(feedback)
}
