// Package nfc provides the DTO for the NFC
package nfc

// CreateNFCRequest 创建NFC请求
type CreateNFCRequest struct {
	DeviceID string `json:"device_id" validate:"required" msg:"required:设备ID不能为空"`
	Ctype    string `json:"ctype" form:"ctype" validate:"required,oneof=明信片 生日卡片 自定义 每日鼓励 悄悄话 成长日记"`
	Title    string `json:"title" form:"title" validate:"required,min=1,max=8" msg:"required:标题不能为空|min:标题不能为空|max:标题不能超过8个字符"`
	Content  string `json:"content" form:"content" validate:"required,max=50" msg:"required:内容不能为空|max:内容不能超过50个字符"`
	Voice    string `json:"voice" form:"voice"`
	Picture  string `json:"picture" form:"picture"`
	Dur      *int   `json:"dur" form:"dur" validate:"omitempty,max=60" msg:"max:语音时长最长60s"`
}

// CreateNFCResponse 创建NFC响应
type CreateNFCResponse struct {
	NFCID string `json:"nfc_id"`
}

// GetNFCInfoRequest 获取NFC信息请求
type GetNFCInfoRequest struct {
	NFCID string `json:"nfc_id" param:"nfc_id" validate:"required"`
}

// GetNFCInfoResponse 获取NFC信息响应
type GetNFCInfoResponse struct {
	NFCID   string `json:"nfc_id"`
	CID     string `json:"cid"`
	Ctype   string `json:"ctype"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// GetNFCListRequest 获取NFC列表请求
type GetNFCListRequest struct {
	DeviceID string `param:"device_id" validate:"required"`
	Ctype    string `json:"ctype" query:"ctype" validate:"omitempty,oneof=明信片 生日卡片 自定义 每日鼓励 悄悄话 成长日记"`
	UpdateAt string `json:"update_at" query:"update_at" validate:"omitempty"`
	Dur      *int   `json:"dur" query:"dur" validate:"omitempty"`
	Page     int    `query:"page" validate:"gte=1" default:"1"`
	PageSize int    `query:"page_size" validate:"gte=1" default:"10"`
}

// ListItem 列表项
type ListItem struct {
	CID       string `json:"cid"`
	Ctype     string `json:"ctype"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Voice     string `json:"voice"`
	Picture   string `json:"picture"`
	Dur       int    `json:"dur"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// GetNFCListResponse 获取NFC列表响应
type GetNFCListResponse struct {
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
	Total    int64      `json:"total"`
	List     []ListItem `json:"list"`
}

// UpdateNFCRequest 更新NFC请求
type UpdateNFCRequest struct {
	CID     string `param:"cid" validate:"required" msg:"required:CID不能为空"`
	Ctype   string `json:"ctype" form:"ctype" validate:"required,oneof==明信片 生日卡片 自定义 每日鼓励 悄悄话 成长日记"`
	Title   string `json:"title" form:"title" validate:"required,max=8" msg:"required:标题不能为空|max:8:标题不能超过8个字符"`
	Content string `json:"content" form:"content" validate:"required,max=50" msg:"required:内容不能为空|max:50:内容不能超过50个字符"`
	Voice   string `json:"voice" form:"voice"`
	Picture string `json:"picture" form:"picture"`
	Dur     *int   `json:"dur" form:"dur" validate:"omitempty,max=60" msg:"max:语音时长最长60s"`
}

// DeleteNFCRequest 删除NFC请求
type DeleteNFCRequest struct {
	CID string `param:"cid" query:"cid" validate:"required"`
}
