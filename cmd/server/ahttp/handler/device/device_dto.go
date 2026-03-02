package devicehandler

// FirstOnlineRequest 设备第一次上线请求
type FirstOnlineRequest struct {
	DeviceID string `json:"device_id" form:"device_id" param:"device_id" query:"device_id" validate:"required,mac"`
	ICCID    string `json:"iccid" form:"iccid" param:"iccid" query:"iccid" validate:"required,min=20,max=30" msg:"required:ICCID不能为空|min:ICCID长度不能小于20|max:ICCID长度不能大于30"`
}

// FirstOnlineResponse 设备第一次上线响应
type FirstOnlineResponse struct {
	MQTTURL      string `json:"mqtt_url"`
	InstanceID   string `json:"instance_id"`
	MQTTUsername string `json:"mqtt_username"`
	MQTTPassword string `json:"mqtt_password"`
}

// BindDeviceRequest 硬件设备发起绑定设备请求
type BindDeviceRequest struct {
	DeviceID string `json:"device_id" form:"device_id" param:"device_id" query:"device_id" validate:"required,mac"`
}
