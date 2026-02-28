// Package ahttp 提供 HTTP 服务相关功能
package ahttp

import (
	"aibuddy/pkg/ahttp"
	"aibuddy/pkg/config"
	"context"
	"fmt"
	"log/slog"
	"net"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

// isEchoInternalRoute 检查是否为 Echo 内部路由
func isEchoInternalRoute(method string) bool {
	return strings.HasPrefix(method, "echo_")
}

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
	base := ahttp.NewBase(router, nil)

	RegisterRoutes(base)

	fmt.Println("\n🚀 Routes registered:")
	fmt.Println("  Method   │ Path")
	fmt.Println("  ─────────┼───────────────────")
	for _, r := range router.Routes() {
		// 过滤掉 Echo 内部路由（以 echo_ 开头的 Method）
		if r.Path != "/*" && !isEchoInternalRoute(r.Method) {
			fmt.Printf("  %-8s │ %s\n", r.Method, r.Path)
		}
	}
	fmt.Println("")

	address := net.JoinHostPort(host, port)

	return router.Start(address)
}

// UniversalMiddlewares 返回通用中间件
func UniversalMiddlewares() []echo.MiddlewareFunc {
	return []echo.MiddlewareFunc{
		middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
			LogMethod:  true,
			LogURI:     true,
			LogStatus:  true,
			LogLatency: true,
			LogError:   true,
			LogHeaders: []string{"Content-Type", "Authorization"},
			LogValuesFunc: func(_ echo.Context, values middleware.RequestLoggerValues) error {
				valuesString := fmt.Sprintf("method=%s, uri=%s, status=%d, latency=%s, headers=%s",
					values.Method,
					values.URI,
					values.Status,
					values.Latency,
					values.Headers,
				)
				switch {
				case values.Status >= 500:
					slog.Error("HTTP: " + valuesString)
				case values.Status >= 400:
					slog.Warn("HTTP: " + valuesString)
				default:
					slog.Info("HTTP: " + valuesString)
				}
				return nil
			},
		}),
		middleware.RecoverWithConfig(middleware.RecoverConfig{
			DisableStackAll:     false,
			DisableErrorHandler: true,
			LogErrorFunc: func(_ echo.Context, err error, stack []byte) error {
				fmt.Printf("%+v \n %+v\n", err, string(stack))

				return nil
			},
		}),
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
		slog.Error("HTTP Error", "error", err, "path", c.Request().URL.Path)
		if err := c.String(500, "Internal Server Error"); err != nil {
			slog.Error("HTTP Error Handler", "error", err)
		}
	}

	router.Validator = &ahttp.Validator{
		Validator: validator.New(),
	}
}
