// Package http 提供 HTTP 服务相关功能
package http

import (
	"aibuddy/cmd/server/http/auth"
	"aibuddy/pkg/config"
	"aibuddy/pkg/http"
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

// StartHTTPServer 启动 HTTP 服务器
func StartHTTPServer(ctx context.Context) error {
	router := echo.New()
	go func() {
		<-ctx.Done()
		if err := router.Shutdown(ctx); err != nil {
			slog.Error("Failed to shutdown HTTP server", "error", err)
		}
	}()

	SetupRoutes(router)
	port := config.Instance.App.Port
	host := config.Instance.App.Host
	base := http.NewBase(router, nil)

	auth.RegisterRoutes(base)

	address := net.JoinHostPort(host, port)

	return router.Start(address)
}

// UniversalMiddlewares 返回通用中间件
func UniversalMiddlewares() []echo.MiddlewareFunc {
	return []echo.MiddlewareFunc{
		middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
			LogURI:     true,
			LogStatus:  true,
			LogLatency: true,
			LogError:   true,
			LogHeaders: []string{"Content-Type", "Authorization"},
			LogValuesFunc: func(c echo.Context, values middleware.RequestLoggerValues) error {
				valuesString := fmt.Sprintf("time=%s, method=%s, uri=%s, status=%d, latency=%s, error=%s, headers=%s",
					time.Now().Format(time.DateTime),
					values.Method,
					values.URI,
					values.Status,
					values.Latency,
					values.Error,
					values.Headers,
				)
				switch {
				case values.Status >= 500:
					c.Logger().Error(valuesString)
				case values.Status >= 400:
					c.Logger().Warn(valuesString)
				default:
					c.Logger().Info(valuesString)
				}
				return nil
			},
		}),
		middleware.RecoverWithConfig(middleware.RecoverConfig{
			DisableStackAll:     false,
			DisableErrorHandler: true,
			LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
				fmt.Println(err, string(stack))
				return nil
			},
		}),
		middleware.Recover(),
		middleware.CORS(),
		middleware.Secure(),
		middleware.BodyLimit("10M"),
		otelecho.Middleware(config.Instance.Tracer.ServiceName),
	}
}

// SetupRoutes 设置路由
func SetupRoutes(router *echo.Echo) {
	router.Use(UniversalMiddlewares()...)

	router.RouteNotFound("/*", func(c echo.Context) error {
		return c.String(404, "404 Not Found")
	})

	router.HideBanner = true
	router.HTTPErrorHandler = func(err error, c echo.Context) {
		c.Logger().Error(err.Error())
		if err := c.String(500, "Internal Server Error"); err != nil {
			slog.Error("HTTP Error Handler", "error", err)
		}
	}

	router.Validator = &http.Validator{
		Validator: validator.New(),
	}
}
