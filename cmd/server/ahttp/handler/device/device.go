// Package devicehandler provides a device handler.
package devicehandler

import (
	"aibuddy/internal/services/device"
	"aibuddy/pkg/ahttp"
	"aibuddy/pkg/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = func() trace.Tracer {
	return otel.Tracer(config.Instance.Tracer.ServiceName)
}

// Device is a device handler.
type Device struct {
	Service *device.Service
}

// NewDevice creates a new device handler.
func NewDevice() *Device {
	return &Device{
		Service: device.NewService(),
	}
}

// FirstOnline 设备第一次上线返回设备的配置信息，如 mqtt 的连接信息
func (d *Device) FirstOnline(state *ahttp.State, req *FirstOnlineRequest) error {
	ctx, span := tracer().Start(state.Context(), "Device.FirstOnline")
	defer span.End()

	configInfo, err := d.Service.FirstOnline(ctx, req.DeviceID, req.ICCID, req.Version)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", req.DeviceID))
		return state.Resposne().Error(err)
	}

	return state.Resposne().Success(&FirstOnlineResponse{
		MQTTURL:      configInfo.MQTTURL,
		InstanceID:   configInfo.InstanceID,
		MQTTUsername: configInfo.MQTTUsername,
		MQTTPassword: configInfo.MQTTPassword,
	})
}

// BindDevice 硬件设备发起绑定设备请求
func (d *Device) BindDevice(state *ahttp.State, _ *BindDeviceRequest) error {
	// ctx, span := tracer().Start(state.Context(), "Device.BindDevice")
	// defer span.End()

	// err := d.Service.BindDevice(ctx, req.DeviceID, req.ICCID)
	// if err != nil {
	// 	span.RecordError(err)
	// 	span.SetAttributes(attribute.String("device_id", req.DeviceID))
	// 	return state.Resposne().Error(err)
	// }
	return state.Resposne().Success()
}
