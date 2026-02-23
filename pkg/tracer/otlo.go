// Package tracer 提供分布式追踪功能
package tracer

import (
	"context"
	"errors"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

// Setup 初始化追踪器
func Setup(serviceName, endpoint string) (*trace.TracerProvider, error) {
	if serviceName == "" || endpoint == "" {
		return nil, errors.New("serviceName or endpoint is required")
	}
	exporter, err := otlptracegrpc.New(context.Background(),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(endpoint),
	)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)

	otel.SetTracerProvider(tp)

	return tp, nil
}

// Teardown 清理追踪器
func Teardown(tp *trace.TracerProvider) error {
	if err := tp.Shutdown(context.Background()); err != nil {
		slog.Error("Failed to shutdown tracer provider", "error", err)
		return err
	}
	slog.Info("Tracer provider shutdown successfully")
	return nil
}
