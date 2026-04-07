// Package emotion is the dto for the emotion trigger
package emotion

import "aibuddy/internal/model"

// GetEmotionsRequest is the dto for the emotion trigger
type GetEmotionsRequest struct {
	DeviceID string `json:"device_id" form:"device_id" param:"device_id" validate:"required" msg:"device_id:设备ID不能为空"`
	Page     int    `json:"page" form:"page" query:"page" validate:"required,min=1" msg:"page:页码不能小于1"`
	PageSize int    `json:"page_size" form:"page_size" query:"page_size" validate:"required,min=1" msg:"page_size:每页数量不能小于1"`
}

// GetEmotionsRes is the response for the emotion trigger
type GetEmotionsRes struct {
	Data  []*model.Emotion `json:"data"`
	Total int64            `json:"total"`
}

// GetLatestEmotionRequest is the request for the latest emotion trigger
type GetLatestEmotionRequest struct {
	DeviceID string `json:"device_id" form:"device_id" param:"device_id" validate:"required" msg:"device_id:设备ID不能为空"`
}

// GetLatestEmotionRes is the response for the latest emotion trigger
type GetLatestEmotionRes struct {
	Data *model.Emotion `json:"data"`
}

// MarkEmotionReadRequest 标记情绪预警已读请求
type MarkEmotionReadRequest struct {
	EmotionIDs []string `json:"emotion_ids" form:"emotion_ids" param:"emotion_ids" query:"emotion_ids" validate:"required" msg:"required:情绪预警ID不能为空"`
}
