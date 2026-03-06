package filehandler

import "mime/multipart"

// UploadFileRequest 上传文件请求
type UploadFileRequest struct {
	DeviceID string                `json:"device_id" param:"device_id" validate:"required,aimac" msg:"required:设备ID不能为空|aimac:设备ID格式无效"`
	File     *multipart.FileHeader `json:"file" form:"file" validate:"required" msg:"required:文件不能为空"`
}

// UploadFileResponse 上传文件响应
type UploadFileResponse struct {
	Filename     string `json:"filename"`
	PresignedURL string `json:"presigned_url"`
}

// FileProxyRequest 文件代理请求
type FileProxyRequest struct {
	DeviceID string `json:"device_id" param:"device_id" validate:"required,aimac" msg:"required:设备ID不能为空|aimac:设备ID格式无效"`
	Filename string `json:"filename" query:"filename" validate:"required" msg:"required:文件名不能为空"`
}

// FileProxyResponse 文件代理响应
type FileProxyResponse struct {
	File []byte `json:"file"`
}
