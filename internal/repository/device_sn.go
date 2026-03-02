package repository

import (
	"aibuddy/internal/query"
	"context"
	"errors"
)

// BindDeviceSnRepo 绑定设备 SN 仓库
type BindDeviceSnRepo struct {
}

// NewBindDeviceSnRepo 创建绑定设备 SN 仓库
func NewBindDeviceSnRepo() *BindDeviceSnRepo {
	return &BindDeviceSnRepo{}
}

// GetDeviceSnByDeviceID 根据设备 ID 获取设备 SN
func (r *BindDeviceSnRepo) GetDeviceSnByDeviceID(ctx context.Context, deviceID string) (string, error) {
	deviceSN, err := query.DeviceSN.WithContext(ctx).Where(query.DeviceSN.DeviceID.Eq(deviceID)).First()
	if err != nil {
		return "", err
	}

	if deviceSN == nil {
		return "", errors.New("device sn not found")
	}

	return deviceSN.SN, nil
}
