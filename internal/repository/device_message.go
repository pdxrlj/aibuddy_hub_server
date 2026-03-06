// Package repository defines the repository for the device message.
package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"context"
	"time"
)

// DeviceMessageRepo 设备消息仓库
type DeviceMessageRepo struct {
}

// NewDeviceMessageRepo 创建设备消息仓库
func NewDeviceMessageRepo() *DeviceMessageRepo {
	return &DeviceMessageRepo{}
}

// CreateDeviceMessage 创建设备消息
func (r *DeviceMessageRepo) CreateDeviceMessage(ctx context.Context, message *model.DeviceMessage) error {
	err := query.DeviceMessage.WithContext(ctx).Create(message)
	if err != nil {
		return err
	}
	return nil
}

// MarkMessageRead 标记消息已读
func (r *DeviceMessageRepo) MarkMessageRead(ctx context.Context, messageID string) error {
	_, err := query.DeviceMessage.WithContext(ctx).Where(query.DeviceMessage.MsgID.Eq(messageID)).
		Updates(map[string]any{
			query.DeviceMessage.UpdatedAt.ColumnName().String(): time.Now(),
			query.DeviceMessage.Read.ColumnName().String():      true,
		})
	if err != nil {
		return err
	}
	return nil
}
