// Package handler 初始化状态信息处理器
package handler

import (
	"aibuddy/aiframe/child"
	"aibuddy/internal/repository"
	"aibuddy/pkg/mqtt"
	"context"
	"log/slog"
	"time"
)

// InitInfoHandler 初始化状态信息处理器
type InitInfoHandler struct {
	DeviceRepo   *repository.DeviceRepo
	DeviceSnRepo *repository.BindDeviceSnRepo
}

// NewInitInfoHandler 创建初始化状态信息处理器
func NewInitInfoHandler() *InitInfoHandler {
	return &InitInfoHandler{
		DeviceRepo:   repository.NewDeviceRepo(),
		DeviceSnRepo: repository.NewBindDeviceSnRepo(),
	}
}

// Handle 处理初始化状态信息
func (h *InitInfoHandler) Handle(ctx *mqtt.Context) {
	defer ctx.Message.Ack()
	deviceID := ctx.Params["device_id"]
	slog.Info("[MQTT] Init", "device_id", deviceID)
	// 1. 发送设备用户信息到设备
	h.DeviceUserInfo(deviceID)
}

// DeviceUserInfo 发送设备用户信息到设备
func (h *InitInfoHandler) DeviceUserInfo(deviceID string) {
	userInfo, err := h.DeviceRepo.GetDeviceInfo(context.Background(), deviceID)
	if err != nil {
		slog.Error("[MQTT] InitInfoHandler get device info failed", "device_id", deviceID, "error", err)
		return
	}

	if userInfo == nil {
		slog.Error("[MQTT] InitInfoHandler user info not found", "device_id", deviceID)
		return
	}

	sn, err := h.DeviceSnRepo.GetDeviceSnByDeviceID(context.Background(), deviceID)
	if err != nil {
		slog.Error("[MQTT] InitInfoHandler get device sn failed", "device_id", deviceID, "error", err)
		return
	}

	if sn == "" {
		slog.Error("[MQTT] InitInfoHandler device sn not found", "device_id", deviceID)
		return
	}

	if err := child.SendChildInfoToDevice(context.Background(), deviceID, &child.Info{
		Sn:       sn,
		NickName: userInfo.DeviceInfo.NickName,
		Sex:      userInfo.DeviceInfo.Gender,
		Birthday: userInfo.DeviceInfo.Birthday.Format(time.DateOnly),
	}); err != nil {
		slog.Error("[MQTT] InitInfoHandler send child info to device failed", "device_id", deviceID, "error", err)
		return
	}

	slog.Info("[MQTT] InitInfoHandler send child info to device success", "device_id", deviceID)
}
