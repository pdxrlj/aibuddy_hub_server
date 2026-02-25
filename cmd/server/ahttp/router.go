// Package ahttp 提供应用层 HTTP 路由配置
package ahttp

import (
	"aibuddy/cmd/server/ahttp/userhandler"
	"aibuddy/pkg/ahttp"
)

// RegisterRoutes 注册认证路由
func RegisterRoutes(base *ahttp.Base) {
	user := userhandler.New()
	base.POST("/login", user.Login)
}
