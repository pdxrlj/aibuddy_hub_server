// Package adebug 调试工具
package adebug

// ParseTokenRequest 解析Token请求
type ParseTokenRequest struct {
	Token string `json:"token" form:"token" query:"token"`
}

// ParseTokenResponse 解析Token响应
type ParseTokenResponse struct {
	UID    int64  `json:"uid"`
	Phone  string `json:"phone"`
	OpenID string `json:"open_id"`
	Exp    string `json:"exp"`
	Iat    string `json:"iat"`
}
