// Package ahttp 提供应用层 HTTP 路由配置
package ahttp

import (
	userhandler "aibuddy/cmd/server/ahttp/handler/user"
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
	base.POST("/login", h.Login)
	base.Group("/user", func(group *ahttp.Group) {
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
}
