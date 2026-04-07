// Package voiceclone 语音克隆
package voiceclone

import (
	"aibuddy/internal/repository"
	"aibuddy/pkg/config"
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = func() trace.Tracer {
	return otel.Tracer(config.Instance.Tracer.ServiceName)
}

// VoiceClone 语音克隆
type VoiceClone struct {
	DeviceRepo *repository.DeviceRepo
}

// NewVoiceClone 实例化语音克隆
func NewVoiceClone() *VoiceClone {
	return &VoiceClone{
		DeviceRepo: repository.NewDeviceRepo(),
	}
}

// SetDeviceVoiceID 设置设备语音ID
func (v *VoiceClone) SetDeviceVoiceID(ctx context.Context, deviceID string, voiceID string) error {
	_, span := tracer().Start(ctx, "SetDeviceVoiceID")
	defer span.End()
	span.SetAttributes(attribute.String("device_id", deviceID))
	span.SetAttributes(attribute.String("voice_id", voiceID))

	return v.DeviceRepo.SetVoiceID(ctx, deviceID, voiceID)
}
