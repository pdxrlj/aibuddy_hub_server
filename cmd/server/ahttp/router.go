// Package ahttp 提供应用层 HTTP 路由配置
package ahttp

import (
	devicehandler "aibuddy/cmd/server/ahttp/handler/device"
	userhandler "aibuddy/cmd/server/ahttp/handler/user"
	"aibuddy/cmd/server/ahttp/middleware"
	"aibuddy/pkg/ahttp"
	"time"
)

// DemmoRequest 演示请求结构
type DemmoRequest struct {
	Name string `json:"name" validate:"required,min=3,max=10"`
	Age  int    `json:"age" validate:"required,min=18,max=100"`
}

// RegisterRoutes 注册认证路由
func RegisterRoutes(base *ahttp.Base) {
	h := userhandler.New()

	base.Group("/api/v1", nil, func(group *ahttp.Group) {
		group.Group("/device", nil, func(deviceGroup *ahttp.Group) {
			device := devicehandler.NewDevice()
			deviceGroup.GET("/firstonline", device.FirstOnline)
		})
	})

	base.POST("/login", h.Login)
	base.Group("/user", nil, func(group *ahttp.Group) {
		group.POST("/info", func(state *ahttp.State, request *DemmoRequest) error {
			return state.Resposne().Success(request)
		})

		group.POST("/info2", func(state *ahttp.State, request DemmoRequest) error {
			return state.Resposne().Success(request)
		})

		group.GET("/time", func(state *ahttp.State) error {
			return state.Resposne().Success(time.Now())
		})
	})

	base.POST("/phone_login", h.PhoneLogin, middleware.UnifiedAuthMiddleware())
}
