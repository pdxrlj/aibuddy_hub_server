// Package repository is the repository for the emotion trigger
package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"context"
	"strconv"
	"time"

	"gorm.io/gen"
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

// GetUnreadCount 获取未读情绪预警数量
func (r *EmotionRepo) GetUnreadCount(ctx context.Context, deviceID string) (int64, error) {
	return query.Emotion.WithContext(ctx).
		Where(query.Emotion.DeviceID.Eq(deviceID)).
		Where(query.Emotion.Read.Is(false)).
		Count()
}

// MarkEmotionRead 标记情绪预警已读
func (r *EmotionRepo) MarkEmotionRead(ctx context.Context, emotionIDs []string) error {
	if len(emotionIDs) == 0 {
		return nil
	}
	ids := make([]int64, 0, len(emotionIDs))
	for _, idStr := range emotionIDs {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return nil
	}

	_, err := query.Emotion.WithContext(ctx).
		Where(query.Emotion.ID.In(ids...)).
		Updates(map[string]any{
			query.Emotion.UpdatedAt.ColumnName().String(): time.Now(),
			query.Emotion.Read.ColumnName().String():      true,
		})
	return err
}

// GetLatestEmotion 获取最新的
func (r *EmotionRepo) GetLatestEmotion(ctx context.Context, deviceID string) (*model.Emotion, error) {
	return query.Emotion.WithContext(ctx).
		Where(query.Emotion.DeviceID.Eq(deviceID)).
		Order(query.Emotion.CreatedAt.Desc()).
		First()
}

// GetEmotionsByDeviceID 获取设备在指定时间范围内的情绪列表
func (r *EmotionRepo) GetEmotionsByDeviceID(ctx context.Context, deviceID string, startTime, endTime time.Time, confidences ...float64) ([]*model.Emotion, error) {
	return query.Emotion.WithContext(ctx).
		Scopes(func(d gen.Dao) gen.Dao {
			if len(confidences) > 0 {
				// 必须是大于等于confidences[0]的值
				return d.Where(query.Emotion.Confidence.Gte(confidences[0]))
			}
			return d
		}).
		Where(query.Emotion.DeviceID.Eq(deviceID)).
		Where(query.Emotion.CreatedAt.Between(startTime, endTime)).
		Order(query.Emotion.CreatedAt.Desc()).
		Find()
}
