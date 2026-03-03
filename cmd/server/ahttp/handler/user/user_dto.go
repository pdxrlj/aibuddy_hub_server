// Package userhandler 提供用户相关的 HTTP 处理器
package userhandler

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// NewLoginRequest 微信登录请求
type NewLoginRequest struct {
	Source        string `json:"source" validate:"required" msg:"required:选择登录方式"`  // 登录来源（mini=小程序,phone=手机号）
	WechatCode    string `json:"wechat_code" validate:"required_if=Source mini"`    // 微信登录临时 code
	EncryptedData string `json:"encrypted_data" validate:"required_if=Source mini"` // 微信加密数据
	IV            string `json:"iv" validate:"required_if=Source mini"`             // 微信加密数据的初始向量

	Phone     string `json:"phone" validate:"required_if=Source phone,chmobile"` // 手机号（仅 source=phone 时必填）
	PhoneCode string `json:"phone_code" validate:"required_if=Source phone"`     // 手机验证码（仅 source=phone 时必填）
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token    string `json:"token"`
	Expires  int64  `json:"expires"`
	UID      int64  `json:"uid"`
	OpenID   string `json:"open_id,omitempty"`
	Nickname string `json:"nickname,omitempty"`
	Phone    string `json:"phone,omitempty"`
}

// SendCodeRequest 验证码请求
type SendCodeRequest struct {
	Phone string `json:"phone" validate:"required,chmobile" msg:"required:用户手机号码不能为空|chmobile:手机号格式无效"`
}

// TokenRequest 退出登录请求
type TokenRequest struct {
	Token string `json:"token" validate:"required"`
}

// TokenResponse Token请求响应
type TokenResponse struct {
	Token   string `json:"token"`
	Expires int64  `json:"expires"`
}

// UserinfoRequest 用户信息请求数据
type UserinfoRequest struct {
	ID        int64  `json:"id"`
	BoardType string `json:"board_type" validate:"required" msg:"required:板子类型不能为空"`
	Version   string `json:"version" validate:"required" msg:"required:版本号不能为空"`

	DeviceID string `json:"device_id" validate:"required"`
	NickName string `json:"nickname" validate:"required"`
	Avatar   string `json:"avatar"`
	Gender   int8   `json:"gender" validate:"required,oneof=0 1 2"`
	Birthday string `json:"birthday" validate:"required"`
	Relation string `json:"relation"`

	Hobbies     string `json:"hobbies"`
	Values      string `json:"values"`
	Skills      string `json:"skills"`
	Personality string `json:"personality"`
}
