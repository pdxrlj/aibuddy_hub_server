// Package otahandler provides a ota handler.
package otahandler

import (
	"aibuddy/internal/services/ota"
	"aibuddy/pkg/ahttp"
	"aibuddy/pkg/config"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = func() trace.Tracer {
	return otel.Tracer(config.Instance.Tracer.ServiceName)
}

// Ota OTA处理器
type Ota struct {
	Service *ota.Ota
}

// NewOta 创建OTA处理器
func NewOta() *Ota {
	return &Ota{
		Service: ota.NewOta(),
	}
}

// SendToDevice 发送OTA更新到设备
func (o *Ota) SendToDevice(state *ahttp.State, req *SendToDeviceRequest) error {
	ctx, span := tracer().Start(state.Context(), "Ota.SendToDevice")
	defer span.End()

	err := o.Service.SendToDevice(ctx, &ota.Request{
		DeviceIDs: req.DeviceIDs,
		SendAll:   req.SendAll,

		Version:     req.Version,
		OtaURL:      req.OtaURL,
		ModelURL:    req.ModelURL,
		ResourceURL: req.ResourceURL,
		ForceUpdate: req.ForceUpdate,
	})
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_ids", strings.Join(req.DeviceIDs, ",")))
		return state.Response().Error(err)
	}

	return state.Response().Success()
}
