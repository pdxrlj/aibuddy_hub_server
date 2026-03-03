// Package otahandler provides a ota handler.
package otahandler

// SendToDeviceRequest 发送OTA更新到设备请求
type SendToDeviceRequest struct {
	DeviceIDs []string `json:"device_ids" form:"device_ids" validate:"required_without=SendAll,aimac" msg:"required_without:device_ids不能为空|aimac:device_ids格式错误"`
	SendAll   bool     `json:"send_all" form:"send_all" validate:"required_without=DeviceIDs,boolean" msg:"required_without:send_all不能为空|boolean:send_all格式错误"`

	Version     string `json:"version" form:"version" validate:"required" msg:"required:version不能为空"`
	OtaURL      string `json:"ota_url" form:"ota_url" validate:"required_without_all=ModelURL ResourceURL" msg:"required_without_all:至少需要填写一个URL"`
	ModelURL    string `json:"model_url" form:"model_url" validate:"required_without_all=OtaURL ResourceURL" msg:"required_without_all:至少需要填写一个URL"`
	ResourceURL string `json:"resource_url" form:"resource_url" validate:"required_without_all=OtaURL ModelURL" msg:"required_without_all:至少需要填写一个URL"`
	ForceUpdate bool   `json:"force_update" form:"force_update"`
}
