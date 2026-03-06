// Package nfc provides the DTO for the NFC
package nfc

// CreateNFCRequest 创建NFC请求
type CreateNFCRequest struct {
	Ctype   string `json:"ctype" form:"ctype" validate:"required,oneof=明信片 生日卡片 自定义"`
	Title   string `json:"title" form:"title" validate:"required,max=255"`
	Content string `json:"content" form:"content" validate:"required,max=1024"`
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
