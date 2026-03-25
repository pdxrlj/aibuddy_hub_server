// Package websocket provides a websocket handler.
package websocket

// HandleConnectRequest 处理连接请求
type HandleConnectRequest struct {
	Token string `json:"token" query:"token" validate:"required" msg:"required:token不能为空"`
}
