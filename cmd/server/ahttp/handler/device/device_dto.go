package devicehandler

// FirstOnlineRequest 设备第一次上线请求
type FirstOnlineRequest struct {
	DeviceID string `json:"device_id" form:"device_id" param:"device_id" query:"device_id" validate:"required,mac"`
	SIMCard  string `json:"sim_card" form:"sim_card" param:"sim_card" query:"sim_card" validate:"required,min=20,max=300" msg:"required:SIM卡号不能为空|min:SIM卡号长度不能小于20|max:SIM卡号长度不能大于300"`
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

// GetFriendsRequest 获取好友列表请求
type GetFriendsRequest struct {
	DeviceID string `json:"device_id" param:"device_id" validate:"required,aimac"`
	Page     int    `json:"page" query:"page" validate:"required,min=1"`
	Size     int    `json:"size" query:"size" validate:"required,min=1,max=100"`
}

// GetFriendsResponse 获取好友列表响应
type GetFriendsResponse struct {
	Total   int64                     `json:"total"`
	Page    int                       `json:"page"`
	Size    int                       `json:"size"`
	Friends []*GetFriendsResponseItem `json:"friends"`
}

// GetFriendsResponseItem 获取好友列表响应项
type GetFriendsResponseItem struct {
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
	Sex        string `json:"sex"`
	Avatar     string `json:"avatar"`
	Relation   string `json:"relation"`
}

// GetDeviceInfoRequest 获取设备信息请求
type GetDeviceInfoRequest struct {
	DeviceID       string `json:"device_id" param:"device_id" validate:"required,aimac"`
	TargetDeviceID string `json:"target_device_id" param:"target_device_id" query:"target_device_id" validate:"required,aimac"`
}

// GetDeviceInfoResponse 获取设备信息响应
type GetDeviceInfoResponse struct {
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
	Avatar     string `json:"avatar"`
	Relation   string `json:"relation"`
}

// AddFriendRequest 添加好友请求
type AddFriendRequest struct {
	DeviceID       string `json:"device_id" param:"device_id" validate:"required,aimac" msg:"required:设备ID不能为空|aimac:设备ID格式无效"`
	TargetDeviceID string `json:"target_device_id" form:"target_device_id" validate:"required,aimac,nefield=DeviceID" msg:"required:目标设备ID不能为空|aimac:目标设备ID格式无效|nefield:目标设备ID不能与设备ID相同"`
}

// AddFriendResponse 添加好友响应
type AddFriendResponse struct {
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
	DeviceID string `json:"device_id"`
}

// DeleteFriendRequest 删除好友请求
type DeleteFriendRequest struct {
	DeviceID       string `json:"device_id" param:"device_id" validate:"required,aimac" msg:"required:设备ID不能为空|aimac:设备ID格式无效"`
	TargetDeviceID string `json:"target_device_id" form:"target_device_id" validate:"required,aimac" msg:"required:目标设备ID不能为空|aimac:目标设备ID格式无效"`
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	DeviceID       string `json:"device_id" param:"device_id" validate:"required,aimac" msg:"required:设备ID不能为空"`
	TargetDeviceID string `json:"target_device_id" form:"target_device_id" validate:"required" msg:"required:目标设备ID不能为空"`
	Content        string `json:"content" form:"content" validate:"required" msg:"required:消息内容不能为空"`
	Fmt            string `json:"fmt" form:"fmt" validate:"required,oneof=text voice" msg:"required:消息格式不能为空|oneof:消息格式必须为text或voice"`
	Dur            int    `json:"dur" form:"dur" validate:"required_if_gt=Fmt=voice" msg:"required_if_gt:语音消息时长必须大于0"`
}

// MessageListRequest 消息列表请求
type MessageListRequest struct {
	DeviceID string `json:"device_id" param:"device_id" validate:"required,aimac" msg:"required:设备ID不能为空"`
	Page     int    `json:"page" query:"page" validate:"required" msg:"required:page不能为空"`
	Size     int    `json:"size" query:"size" validate:"required" msg:"required:size不能为空"`
}

// MessageListResponse 消息列表请求
type MessageListResponse struct {
	Page  int   `json:"page"`
	Size  int   `json:"size"`
	Total int64 `json:"total"`
	List  any   `json:"list"`
}

// MessageInfo 消息内容数据
type MessageInfo struct {
	Mid      string `json:"msg_id"`
	From     string `json:"from"`
	FromName string `json:"from_name"`
	Fmt      string `json:"fmt"`
	Content  string `json:"content"`
	Dur      int    `json:"dur"`
	Ts       int
}

// AccountInfoRequest 获取硬件的账户消息请求
type AccountInfoRequest struct {
	DeviceID string `json:"device_id" param:"device_id" validate:"required,aimac" msg:"required:设备ID不能为空"`
}

// AccountInfoResponse 获取硬件的账户消息响应
type AccountInfoResponse struct {
	AccountInfo *AccountInfo `json:"account_info"`
}

// AccountInfo 硬件的账户消息
type AccountInfo struct {
	NickName string `json:"nick_name"`
	Sex      string `json:"sex"`
	Birthday string `json:"birthday"`
	Sn       string `json:"sn"`
}

// OtaCheckRequest ota 升级校验请求
type OtaCheckRequest struct {
	DeviceID string `json:"device_id" param:"device_id" validate:"required,aimac" msg:"required:设备ID不能为空"`
	Version  string `json:"version" form:"version" param:"version" query:"version" validate:"required,semver" msg:"required:版本号不能为空|semver:版本号格式无效"`
}
