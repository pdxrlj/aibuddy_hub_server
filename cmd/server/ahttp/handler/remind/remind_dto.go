package remindhandler

import "aibuddy/internal/model"

// AddRemindRequest 添加提醒事件数据
type AddRemindRequest struct {
	ID              int64  `json:"id"`
	ReminderType    int    `json:"reminder_type"`
	ReminderTitle   string `json:"reminder_title" validate:"required"`
	ReminderContent string `json:"reminder_content" validate:"required"`
	ReminderTime    string `json:"reminder_time"`
	DeviceID        string `json:"device_id" validate:"required"`
	RepeatType      string `json:"repeat_type" validate:"required"`
	Status          string `json:"status" validate:"required"`
}

// RemindRequest 提醒事件
type RemindRequest struct {
	ID int64 `json:"id" validate:"required"`
}

// ListReqeust 提醒事件列表数据
type ListReqeust struct {
	DeviceID string `json:"device_id" validate:"required"`
	Page     int64  `json:"page" validate:"required,min=1"`
	Size     int64  `jsons:"size" validate:"required,min=10"`
}

// ListResponse 列表响应
type ListResponse struct {
	Total int64             `json:"total"`
	Page  int64             `json:"page"`
	Size  int64             `json:"size"`
	Data  []*model.Reminder `json:"data"`
}
