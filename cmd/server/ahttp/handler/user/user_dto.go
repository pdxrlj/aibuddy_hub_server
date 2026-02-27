// Package userhandler 提供用户相关的 HTTP 处理器
package userhandler

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// NewLoginRequest 微信登录请求
type NewLoginRequest struct {
	Source        string `json:"source" validate:"required"`                        // 登录来源（mini=小程序,phone=手机号）
	WechatCode    string `json:"wechat_code" validate:"required_if=Source mini"`    // 微信登录临时 code
	EncryptedData string `json:"encrypted_data" validate:"required_if=Source mini"` // 微信加密数据
	IV            string `json:"iv" validate:"required_if=Source mini"`             // 微信加密数据的初始向量

	Phone     string `json:"phone" validate:"required_if=Source phone"`        // 手机号（仅 source=phone 时必填）
	PhoneCode string `json:"phone_code,"  validate:"required_if=Source phone"` // 手机验证码（仅 source=phone 时必填）
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token    string `json:"token"`
	UID      int64  `json:"uid"`
	OpenID   string `json:"open_id,omitempty"`
	Nickname string `json:"nickname,omitempty"`
	Phone    string `json:"phone,omitempty"`
}
