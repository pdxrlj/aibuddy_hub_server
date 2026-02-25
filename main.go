// aibuddy_hub 入口文件
package main

import (
	"aibuddy/cmd/server"
	"aibuddy/cmd/server/ahttp"
	"aibuddy/cmd/server/amqtt"
	"aibuddy/cmd/server/schedule"
	"aibuddy/pkg/config"
	logger "aibuddy/pkg/log"
	"aibuddy/pkg/tracer"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
)

// SignalHandler 处理系统信号
func SignalHandler(cancel context.CancelFunc) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
	close(signals)
	cancel()
}

// Setup 初始化配置和日志
func Setup() error {
	time.FixedZone("PRC", 0)
	config.Setup()
	logger.NewLogger(config.Instance.App.LogLevel, logger.DefaultFileLogger()).Setup()
	_, err := tracer.Setup(config.Instance.Tracer.ServiceName, config.Instance.Tracer.Endpoint)
	if err != nil {
		slog.Error("Failed to setup tracer", "error", err)
		return err
	}
	return nil
}

// Teardown 清理资源
func Teardown() error {
	tp := otel.GetTracerProvider()
	if tp != nil {
		if provider, ok := tp.(*trace.TracerProvider); ok {
			if err := tracer.Teardown(provider); err != nil {
				return err
			}
		}
	}
	return nil
}

func main() {
	if err := Setup(); err != nil {
		slog.Error("Failed to setup", "error", err)
		return
	}
	defer func() {
		if err := Teardown(); err != nil {
			slog.Error("Failed to teardown", "error", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	go SignalHandler(cancel)
	err := server.StartServer(ctx, ahttp.StartHTTPServer, schedule.StartSchedule, amqtt.StartAMQTTServer)
	if err != nil {
		slog.Error("Failed to start server", "error", err)
	}
	<-ctx.Done()
	slog.Info("Server stopped")
}
