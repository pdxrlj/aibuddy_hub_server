// Package repository defines the repository for the device message.
package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"context"
	"strings"
	"time"

	"github.com/spf13/cast"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gen/field"
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
func (r *DeviceMessageRepo) BatchMessageRead(ctx context.Context, messageID []string) error {
	ids := strings.Join(messageID, ",")
	_, err := query.DeviceMessage.WithContext(ctx).Debug().
		Where(
			query.DeviceMessage.MsgID.In(ids),
		).
		Updates(map[string]any{
			query.DeviceMessage.UpdatedAt.ColumnName().String(): time.Now(),
			query.DeviceMessage.Read.ColumnName().String():      true,
		})
	if err != nil {
		return err
	}
	return nil
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

	messages, total, err := query.DeviceMessage.WithContext(ctx).
		Debug().
		Or(query.DeviceMessage.FromDeviceID.Eq(fromID)).
		Or(query.DeviceMessage.ToDeviceID.Eq(fromID)).
		Preload(query.DeviceMessage.Device.DeviceInfo).
		Preload(query.DeviceMessage.ToDevice.DeviceInfo).
		FindByPage(offset, size)

	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("from_id", fromID), attribute.Int("page", page), attribute.Int("size", size))
		return nil, 0, err
	}
	return messages, total, nil
}

// GetConvMessageList 获取两个设备之间的对话消息列表（双向：A->B 和 B->A）
func (r *DeviceMessageRepo) GetConvMessageList(ctx context.Context, deviceID, targetDeviceID string, page, size int) ([]*model.DeviceMessage, int64, error) {
	_, span := tracer.Start(ctx, "DeviceMessageRepo.GetConvMessageList")
	defer span.End()

	offset := (page - 1) * size

	// 查询双向消息：(A->B) OR (B->A)
	messages, total, err := query.DeviceMessage.WithContext(ctx).
		Where(
			field.Or(
				// A -> B
				field.And(
					query.DeviceMessage.FromDeviceID.Eq(deviceID),
					query.DeviceMessage.ToDeviceID.Eq(targetDeviceID),
				),
				// B -> A
				field.And(
					query.DeviceMessage.FromDeviceID.Eq(targetDeviceID),
					query.DeviceMessage.ToDeviceID.Eq(deviceID),
				),
			),
		).
		Order(query.DeviceMessage.CreatedAt.Desc()).
		Preload(query.DeviceMessage.Device.DeviceInfo).
		Preload(query.DeviceMessage.ToDevice.DeviceInfo).
		FindByPage(offset, size)

	if err != nil {
		span.RecordError(err)
		span.SetAttributes(
			attribute.String("device_id", deviceID),
			attribute.String("target_device_id", targetDeviceID),
			attribute.Int("page", page),
			attribute.Int("size", size),
		)
		return nil, 0, err
	}

	return messages, total, nil
}

// GetUnreadMessageCount 获取未读消息数量
func (r *DeviceMessageRepo) GetUnreadMessageCount(ctx context.Context, uid int64, deviceID string) (int64, *model.DeviceMessage, error) {
	_, span := tracer.Start(ctx, "DeviceMessageRepo.GetUnreadMessageCount")
	defer span.End()

	uidStr := cast.ToString(uid)
	deviceMessages, err := query.DeviceMessage.WithContext(ctx).
		Debug().
		Order(query.DeviceMessage.ID.Desc()).
		Where(
			field.Or(
				// FromDeviceID=uid AND ToDeviceID=deviceID
				field.And(
					query.DeviceMessage.FromDeviceID.Eq(uidStr),
					query.DeviceMessage.ToDeviceID.Eq(deviceID),
				),
				// FromDeviceID=deviceID AND ToDeviceID=uid
				field.And(
					query.DeviceMessage.FromDeviceID.Eq(deviceID),
					query.DeviceMessage.ToDeviceID.Eq(uidStr),
				),
			),
		).
		Where(query.DeviceMessage.Read.Is(false)).
		Find()
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.Int64("uid", uid), attribute.String("device_id", deviceID))
		return 0, nil, err
	}
	if len(deviceMessages) == 0 {
		return 0, nil, nil
	}
	LastUnreadMsg := deviceMessages[0]

	return int64(len(deviceMessages)), LastUnreadMsg, nil
}
