// Package repository is the repository for the emotion trigger
package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"context"
)

// EmotionRepo is the repository for the emotion
type EmotionRepo struct {
}

// NewEmotionRepo creates a new emotion repository
func NewEmotionRepo() *EmotionRepo {
	return &EmotionRepo{}
}

// CreateEmotion creates a new emotion
func (r *EmotionRepo) CreateEmotion(ctx context.Context, trigger *model.Emotion) error {
	return query.Emotion.WithContext(ctx).Create(trigger)
}

// GetEmotions 获取情绪列表
func (r *EmotionRepo) GetEmotions(ctx context.Context, page, pageSize int, deviceID string) ([]*model.Emotion, int64, error) {
	offset := (page - 1) * pageSize
	return query.Emotion.WithContext(ctx).
		Where(query.Emotion.DeviceID.Eq(deviceID)).
		FindByPage(offset, pageSize)
}

// GetLatestEmotion 获取最新的
func (r *EmotionRepo) GetLatestEmotion(ctx context.Context, deviceID string) (*model.Emotion, error) {
	return query.Emotion.WithContext(ctx).
		Where(query.Emotion.DeviceID.Eq(deviceID)).
		Order(query.Emotion.CreatedAt.Desc()).
		First()
}
