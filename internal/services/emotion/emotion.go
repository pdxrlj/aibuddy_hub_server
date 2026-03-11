// Package emotiontrigger is the service for the emotion trigger
package emotiontrigger

import (
	"aibuddy/internal/model"
	"aibuddy/internal/repository"
	"context"
)

// Service is the service for the emotion
type Service struct {
	emotionRepo *repository.EmotionRepo
}

// NewService creates a new emotion service
func NewService() *Service {
	return &Service{
		emotionRepo: repository.NewEmotionRepo(),
	}
}

// GetEmotions gets the emotions
func (s *Service) GetEmotions(ctx context.Context, page, pageSize int, deviceID string) ([]*model.Emotion, int64, error) {
	return s.emotionRepo.GetEmotions(ctx, page, pageSize, deviceID)
}

// GetLatestEmotion gets the latest emotion
func (s *Service) GetLatestEmotion(ctx context.Context, deviceID string) (*model.Emotion, error) {
	return s.emotionRepo.GetLatestEmotion(ctx, deviceID)
}
