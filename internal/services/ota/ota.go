// Package ota provides a ota service.
package ota

import (
	"aibuddy/aiframe/ota"
	"aibuddy/internal/model"
	"aibuddy/internal/repository"
	"aibuddy/pkg/config"
	"context"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// tracer 获取tracer
var tracer = func() trace.Tracer {
	return otel.Tracer(config.Instance.Tracer.ServiceName)
}

// Ota 提供OTA服务
type Ota struct {
	deviceRepo    *repository.DeviceRepo
	deviceOtaRepo *repository.DeviceOtaRepo
}

// NewOta 创建OTA服务
func NewOta() *Ota {
	return &Ota{
		deviceRepo:    repository.NewDeviceRepo(),
		deviceOtaRepo: repository.NewDeviceOtaRepo(),
	}
}

// Request 发送OTA更新请求
type Request struct {
	DeviceIDs []string `json:"device_ids"`
	SendAll   bool     `json:"send_all"`

	Version     string `json:"version"`
	OtaURL      string `json:"ota_url"`
	ModelURL    string `json:"model_url"`
	ResourceURL string `json:"resource_url"`
	ForceUpdate bool   `json:"force_update"`
}

// SendToDevice 发送OTA更新
func (o *Ota) SendToDevice(ctx context.Context, req *Request) error {
	subCtx, span := tracer().Start(ctx, "Ota.SendToDevice")
	defer span.End()

	if req.SendAll {
		// 发送所有设备
		span.SetAttributes(attribute.Bool("send_all", req.SendAll))
		return o.sendAllDevices(subCtx, req.Version, req.OtaURL, req.ModelURL, req.ResourceURL, req.ForceUpdate)
	}
	// 发送指定设备
	span.SetAttributes(attribute.Bool("send_all", req.SendAll))
	span.SetAttributes(attribute.String("device_ids", strings.Join(req.DeviceIDs, ",")))
	return o.sendSpecifiedDevices(subCtx, req.Version, req.OtaURL, req.ModelURL, req.ResourceURL, req.ForceUpdate, req.DeviceIDs)
}

// sendAllDevices 发送所有设备
func (o *Ota) sendAllDevices(ctx context.Context, version string, otaURL, modelURL, resourceURL string, forceUpdate bool) error {
	subCtx, span := tracer().Start(ctx, "Ota.sendAllDevices")
	defer span.End()

	ota := &ota.Ota{
		Type:        "ota",
		Version:     version,
		OtaURL:      otaURL,
		ModelURL:    modelURL,
		ResourceURL: resourceURL,
		ForceUpdate: forceUpdate,
	}

	return o.deviceRepo.BatchHandlerDeviceList(subCtx, func(devices []*model.Device) {
		for _, device := range devices {
			if err := ota.SendToDevice(device.DeviceID); err != nil {
				span.RecordError(err)
			} else {
				deviceOta := &model.DeviceOta{
					DeviceID:    device.DeviceID,
					Version:     version,
					OtaURL:      otaURL,
					ModelURL:    modelURL,
					ResourceURL: resourceURL,
					ForceUpdate: forceUpdate,
				}

				if err := o.deviceOtaRepo.UpsertDeviceOta(subCtx, deviceOta); err != nil {
					span.RecordError(err)
				}
			}
		}
	})
}

// sendSpecifiedDevices 发送指定设备
func (o *Ota) sendSpecifiedDevices(ctx context.Context, version string, otaURL, modelURL, resourceURL string, forceUpdate bool, deviceIDs []string) error {
	subCtx, span := tracer().Start(ctx, "Ota.sendSpecifiedDevices")
	defer span.End()

	ota := &ota.Ota{
		Type:        "ota",
		Version:     version,
		OtaURL:      otaURL,
		ModelURL:    modelURL,
		ResourceURL: resourceURL,
		ForceUpdate: forceUpdate,
	}

	span.SetAttributes(attribute.String("ota", ota.String()))

	for _, deviceID := range deviceIDs {
		if err := ota.SendToDevice(deviceID); err != nil {
			span.RecordError(err)
		}

		deviceOta := &model.DeviceOta{
			DeviceID:    deviceID,
			Version:     version,
			OtaURL:      otaURL,
			ModelURL:    modelURL,
			ResourceURL: resourceURL,
			ForceUpdate: forceUpdate,
		}

		if err := o.deviceOtaRepo.UpsertDeviceOta(subCtx, deviceOta); err != nil {
			span.RecordError(err)
		}
	}
	return nil
}
