// Package ahttp 提供应用层 HTTP 路由配置
package ahttp

import (
	devicehandler "aibuddy/cmd/server/ahttp/handler/device"
	rolehandleer "aibuddy/cmd/server/ahttp/handler/role"
	userhandler "aibuddy/cmd/server/ahttp/handler/user"
	"aibuddy/cmd/server/ahttp/middleware"
	"aibuddy/pkg/ahttp"

	"github.com/labstack/echo/v4"
)

// DemmoRequest 演示请求结构
type DemmoRequest struct {
	Name string `json:"name" validate:"required,min=3,max=10"`
	Age  int    `json:"age" validate:"required,min=18,max=100"`
}

// RegisterRoutes 注册认证路由
func RegisterRoutes(base *ahttp.Base) {
	base.Group("/api/v1", nil, func(group *ahttp.Group) {
		group.Group("/device", nil, func(deviceGroup *ahttp.Group) {
			device := devicehandler.NewDevice()
			deviceGroup.GET("/firstonline", device.FirstOnline)

			// 硬件设备发起绑定设备请求
			deviceGroup.POST("/bind", device.BindDevice)
		})

		group.Group("/user", []echo.MiddlewareFunc{middleware.UnifiedAuthMiddleware()}, func(userGroup *ahttp.Group) {
			h := userhandler.New()
			userGroup.POST("/send_code", h.SendCode)
			userGroup.POST("/login", h.Login)
			userGroup.POST("/refresh_token", h.RefreshToken)
			userGroup.POST("/logout", h.Logout)

			// 完善用户信息，扫描绑定后完善用户信息
			userGroup.POST("/profile", h.CompleteProfile)
		})

		group.Group("/role", []echo.MiddlewareFunc{middleware.UnifiedAuthMiddleware()}, func(group *ahttp.Group) {
			r := rolehandleer.NewRoleHandler()
			group.GET("/list", r.RoleList)
			group.POST("/change", r.ChangeRole) // 切换角色
		})
	})
}
