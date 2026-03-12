package anniversaryhandler

// AnniversaryInfoRequest  获取列表请求
type AnniversaryInfoRequest struct {
	ID              int64  `json:"id"`
	DeviceID        string `json:"device_id"  validate:"required,aimac"  msg:"required:设备ID不能为空|aimac:设备ID格式无效"`
	AnniversaryType string `json:"anniversary_type" validate:"required" msg:"required:纪念日类型不能为空"`

	ReminderUsername string `json:"reminder_username" validate:"required" msg:"required:提醒用户名称不能为空"`
	ReminderUserSex  string `json:"reminder_user_sex"`
	AnniversaryTime  string `json:"anniversary_time" validate:"required" msg:"required:提醒时间不能为空"`
	ReminderWay      string `json:"reminder_way" validate:"required" msg:"required:提醒方式不能为空"`
	Remarks          string `json:"remarks"`
}

// InfoRequest 指定数据ID请求
type InfoRequest struct {
	ID       int64  `json:"id"  validate:"required" msg:"required:ID不能为空"`
	DeviceID string `json:"device_id"  validate:"required,aimac" msg:"required:设备ID不能为空|aimac:设备ID格式无效"`
}

// ListRequest 列表请求数据
type ListRequest struct {
	DeviceID string `json:"device_id" query:"device_id" validate:"required,aimac" msg:"required:设备ID不能为空|aimac:设备ID格式无效"`
	Page     int    `json:"page" query:"page"`
	Size     int    `json:"size" query:"size" validate:"required" msg:"required:size不能为空"`
}

// ListReponse 列表响应
type ListReponse struct {
	Total int64          `json:"total"`
	Page  int            `json:"page"`
	Size  int            `json:"size"`
	Data  []*InfoReponse `json:"data"`
}

// InfoReponse 列表响应
type InfoReponse struct {
	ID               int64  `json:"id"`
	DeviceID         string `json:"device_id"`
	AnniversaryType  string `json:"anniversary_type"`
	ReminderUsername string `json:"reminder_username"`
	ReminderUserSex  string `json:"reminder_user_sex"`
	AnniversaryTime  string `json:"anniversary_time"`
	ReminderWay      string `json:"reminder_way"`
	Remarks          string `json:"remarks"`
}
