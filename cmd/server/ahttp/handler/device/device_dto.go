package devicehandler

// FirstOnlineRequest 设备第一次上线请求
type FirstOnlineRequest struct {
	DeviceID string `json:"device_id" form:"device_id" param:"device_id" query:"device_id" validate:"required,mac"`
	ICCID    string `json:"iccid" form:"iccid" param:"iccid" query:"iccid" validate:"required,min=20,max=30" msg:"required:ICCID不能为空|min:ICCID长度不能小于20|max:ICCID长度不能大于30"`
	Version  string `json:"version" form:"version" param:"version" query:"version" validate:"required,semver" msg:"required:版本号不能为空|semver:版本号格式无效"`
}

// FirstOnlineResponse 设备第一次上线响应
type FirstOnlineResponse struct {
	MQTTConfig *MQTTConfig `json:"mqtt_config"`

	DeviceInfo *DeviceInfo `json:"device_info"`
}

// MQTTConfig MQTT配置
type MQTTConfig struct {
	MQTTURL      string `json:"mqtt_url"`
	InstanceID   string `json:"instance_id"`
	MQTTUsername string `json:"mqtt_username"`
	MQTTPassword string `json:"mqtt_password"`
}

// DeviceInfo 设备信息
type DeviceInfo struct {
	UserID     string `json:"user_id"`
	InstanceID uint64 `json:"instance_id"`
}

// GetLocationRequest 获取设备位置请求
type GetLocationRequest struct {
	DeviceID string `json:"device_id" form:"device_id" param:"device_id" query:"device_id" validate:"required,aimac"`
}
