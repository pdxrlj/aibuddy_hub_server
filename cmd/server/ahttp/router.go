// Package ahttp 提供应用层 HTTP 路由配置
package ahttp

import (
	"aibuddy/cmd/server/ahttp/userhandler"
	"aibuddy/pkg/ahttp"
	"time"
)

type DemmoRequest struct {
	Name string `json:"name" validate:"required,min=3,max=10"`
	Age  int    `json:"age" validate:"required,min=18,max=100"`
}

// RegisterRoutes 注册认证路由
func RegisterRoutes(base *ahttp.Base) {
	user := userhandler.New()
	base.POST("/login", user.Login)
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
