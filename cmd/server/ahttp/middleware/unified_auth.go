// Package middleware 用户认证中间件
package middleware

import (
	aiuserService "aibuddy/internal/services/aiuser"
	"aibuddy/pkg/ahttp"
	"errors"
	"slices"
	"strings"

	"github.com/labstack/echo/v4"
)

// SkipPaths 跳过认证的路径
var SkipPaths = []string{"/api/v1/user/login", "/api/v1/user/send_code"}

// UnifiedAuthMiddleware 统一认证中间件，支持多种认证方式（如 JWT、API Key 等）
func UnifiedAuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		// cfg := config.Instance.App
		return func(c echo.Context) error {
			// 跳过白名单路径
			if slices.Contains(SkipPaths, c.Path()) {
				return next(c)
			}

			// 1.尝试微信验证
			if token := c.Request().Header.Get("Authorization"); token != "" {
				token = strings.TrimPrefix(token, "Bearer ")
				_, err := aiuserService.ValidateToken(c, token)
				if err == nil {
					// 认证成功，继续处理请求
					return next(c)
				}
			}

			// 2.其他认证方式（如 Token）可以在这里添加
			// if token := c.Request().Header.Get("Token"); token != "" {
			// 	if aiuserService.ValidateSimpleToken(token, cfg.AppSecret) {
			// 		return next(c)
			// 	}
			// }

			// 3.尝试Bearer Token认证（兼容模式）
			// if token := c.Request().Header.Get("Authorization"); token != "" {
			// 	token = strings.TrimPrefix(token, "Bearer ")
			// 	if aiuserService.ValidateSimpleToken(token, cfg.AppSecret) {
			// 		return next(c)
			// 	}
			// }

			return ahttp.NewResponse(c).SetStatus(401).Error(errors.New("invalid token"))
		}
	}
}
