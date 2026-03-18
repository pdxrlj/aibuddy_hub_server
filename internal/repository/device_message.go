// Package repository defines the repository for the device message.
package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"context"
	"strings"
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

// BatchMessageRead 批量已读消息
func (r *DeviceMessageRepo) BatchMessageRead(ctx context.Context, deviceID string, messageID []string) error {
	ids := strings.Join(messageID, ",")
	_, err := query.DeviceMessage.WithContext(ctx).Where(query.DeviceMessage.MsgID.In(ids), query.DeviceMessage.ToDeviceID.Eq(deviceID)).
		Updates(map[string]any{
			query.DeviceMessage.UpdatedAt.ColumnName().String(): time.Now(),
			query.DeviceMessage.Read.ColumnName().String():      true,
		})
	if err != nil {
		return err
	}
	return nil
}

// GetMessageList 消息列表
func (r *DeviceMessageRepo) GetMessageList(ctx context.Context, deviceID string, page int, size int) ([]*model.DeviceMessage, int64, error) {
	offset := (page - 1) * size
	return query.DeviceMessage.WithContext(ctx).
		Where(query.DeviceMessage.ToDeviceID.Eq(deviceID)).
		Or(query.DeviceMessage.FromDeviceID.Eq(deviceID)).
		// Preload()
		FindByPage(offset, size)
}

// GetMessageListByDeviceID 获取设备在指定时间范围内的消息列表
func (r *DeviceMessageRepo) GetMessageListByDeviceID(ctx context.Context, deviceID string, startTime, endTime time.Time) ([]*model.DeviceMessage, error) {
	_, span := tracer.Start(ctx, "DeviceMessageRepo.GetMessageListByDeviceID")
	defer span.End()

	return query.DeviceMessage.WithContext(ctx).
		Where(query.DeviceMessage.ToDeviceID.Eq(deviceID)).
		Or(query.DeviceMessage.FromDeviceID.Eq(deviceID)).
		Where(
			query.DeviceMessage.CreatedAt.Gte(model.LocalTime(startTime)),
			query.DeviceMessage.CreatedAt.Lte(model.LocalTime(endTime)),
		).
		Preload(query.DeviceMessage.Device).
		Preload(query.DeviceMessage.ToDevice).
		Find()
}

// GetMessageFromUser 获取指定用户之间的留言
func (r *DeviceMessageRepo) GetMessageFromUser(ctx context.Context, fromID string, page, size int) ([]*model.DeviceMessage, int64, error) {
	_, span := tracer.Start(ctx, "DeviceMessageRepo.GetMessageFromUser")
	defer span.End()

	offset := (page - 1) * size

	return query.DeviceMessage.WithContext(ctx).
		Debug().
		Or(query.DeviceMessage.FromDeviceID.Eq(fromID)).
		Or(query.DeviceMessage.ToDeviceID.Eq(fromID)).
		Preload(query.DeviceMessage.Device.DeviceInfo).
		Preload(query.DeviceMessage.ToDevice.DeviceInfo).
		FindByPage(offset, size)
}
