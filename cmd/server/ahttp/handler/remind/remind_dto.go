package remindhandler

// AddRemindRequest 添加提醒事件数据
type AddRemindRequest struct {
	ID              int64  `json:"id"`
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
	DeviceID string `json:"device_id" form:"device_id" param:"device_id" query:"device_id"  validate:"required"`
	Page     int64  `json:"page" form:"page" param:"page" query:"page" validate:"required,min=1"`
	Size     int64  `jsons:"size" form:"size" param:"size" query:"size" validate:"required,min=10"`
}

// ListResponse 列表响应
type ListResponse struct {
	Total int64           `json:"total"`
	Page  int64           `json:"page"`
	Size  int64           `json:"size"`
	Data  []*ReminderInfo `json:"data"`
}

// ReminderInfo 提醒事件数据
type ReminderInfo struct {
	ID              int64  `json:"id"`
	DeviceID        string `json:"device_id"`
	RepeatType      string `json:"repeat_type"`
	ReminderTitle   string `json:"reminder_title"`
	ReminderContent string `json:"reminder_content"`
	ReminderTime    string `json:"reminder_time"`
	Status          string `json:"status"`
}
