package filehandler

import "mime/multipart"

// UploadFileRequest 上传文件请求
type UploadFileRequest struct {
	DeviceID string                `json:"device_id" param:"device_id" validate:"required,aimac" msg:"required:设备ID不能为空|aimac:设备ID格式无效"`
	File     *multipart.FileHeader `json:"file" form:"file" validate:"required" msg:"required:文件不能为空"`

	EnableAudioTranscode bool   `json:"enable_audio_transcode" form:"enable_audio_transcode"`
	DestAudioFormat      string `json:"dest_audio_format" form:"dest_audio_format" validate:"required_if=EnableAudioTranscode true,oneof=mp3 wav aac flac ogg opus m4a" msg:"required_if:目标音频格式不能为空|oneof:音频格式无效"`
}

// UploadFileResponse 上传文件响应
type UploadFileResponse struct {
	Filename     string `json:"filename"`
	PresignedURL string `json:"presigned_url"`
}

// UploadStreamRequest 流式上传请求（纯二进制流）
type UploadStreamRequest struct {
	DeviceID string `param:"device_id" validate:"required,aimac" msg:"required:设备ID不能为空|aimac:设备ID格式无效"`
	Ext      string `query:"ext" validate:"required,oneof=.wav .mp3 .pcm .aac .m4a .ogg .flac .opus" msg:"required:文件扩展名不能为空|oneof:不支持的文件格式"`

	EnableAudioTranscode bool   `query:"enable_audio_transcode"`
	DestAudioFormat      string `query:"dest_audio_format" validate:"required_if=EnableAudioTranscode true,oneof=mp3 wav aac flac ogg opus m4a" msg:"required_if:目标音频格式不能为空|oneof:音频格式无效"`
}

// SkipBodyBind 实现 SkipBodyBinder 接口，跳过 body 绑定
func (r *UploadStreamRequest) SkipBodyBind() {}

// FileProxyRequest 文件代理请求
type FileProxyRequest struct {
	DeviceID string `json:"device_id" param:"device_id" validate:"required,aimac" msg:"required:设备ID不能为空|aimac:设备ID格式无效"`
	Filename string `json:"filename" query:"filename" validate:"required" msg:"required:文件名不能为空"`
}

// FileProxyResponse 文件代理响应
type FileProxyResponse struct {
	File []byte `json:"file"`
}
