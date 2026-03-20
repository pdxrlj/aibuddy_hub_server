// Package repository provides a repository for the device relationship.
package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gen/field"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DeviceRelationshipRepo 设备关系仓库
type DeviceRelationshipRepo struct{}

// NewDeviceRelationshipRepo 创建设备关系仓库实例
func NewDeviceRelationshipRepo() *DeviceRelationshipRepo {
	return &DeviceRelationshipRepo{}
}

// GetFriends 获取好友列表 双向验证,双方都同意才算好友
func (d *DeviceRelationshipRepo) GetFriends(ctx context.Context, deviceID string, page, size int) ([]*model.DeviceRelationship, int64, error) {
	_, span := tracer.Start(ctx, "DeviceRelationshipRepo.GetFriends")
	defer span.End()

	offset := (page - 1) * size

	var reverseDeviceIDs []string
	err := query.DeviceRelationship.WithContext(ctx).
		Where(query.DeviceRelationship.TargetDeviceID.Eq(deviceID)).
		Where(query.DeviceRelationship.Status.Eq(model.RelationshipStatusAccepted.String())).
		Pluck(query.DeviceRelationship.DeviceID, &reverseDeviceIDs)
	if err != nil {
		span.RecordError(err)
		return nil, 0, err
	}

	if len(reverseDeviceIDs) == 0 {
		return []*model.DeviceRelationship{}, 0, nil
	}

	relationships, total, err := query.DeviceRelationship.
		Debug().
		Preload(query.DeviceRelationship.Device).
		Preload(query.DeviceRelationship.TargetDevice).
		Preload(query.DeviceRelationship.Device.DeviceInfo).
		Preload(field.NewRelation("TargetDevice.DeviceInfo", "model.DeviceInfo")).
		Where(query.DeviceRelationship.DeviceID.Eq(deviceID)).
		Where(query.DeviceRelationship.Status.Eq(model.RelationshipStatusAccepted.String())).
		Where(query.DeviceRelationship.TargetDeviceID.In(reverseDeviceIDs...)).
		Order(query.DeviceRelationship.CreatedAt.Desc()).
		FindByPage(offset, size)
	if err != nil {
		span.RecordError(err)
		return nil, 0, err
	}

	return relationships, total, nil
}

// IsFriend 判断是否是好友 双向验证，双方都同意才算好友
func (d *DeviceRelationshipRepo) IsFriend(ctx context.Context, deviceID, targetDeviceID string) (bool, error) {
	_, span := tracer.Start(ctx, "DeviceRelationshipRepo.IsFriend")
	defer span.End()

	// 查 deviceID -> targetDeviceID 的关系
	rel1, err := query.DeviceRelationship.
		Where(query.DeviceRelationship.DeviceID.Eq(deviceID)).
		Where(query.DeviceRelationship.TargetDeviceID.Eq(targetDeviceID)).
		Where(query.DeviceRelationship.Status.Eq(model.RelationshipStatusAccepted.String())).
		First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		span.RecordError(err)
		return false, err
	}
	if rel1 == nil {
		return false, nil
	}

	// 查 targetDeviceID -> deviceID 的关系（反向验证）
	rel2, err := query.DeviceRelationship.
		Where(query.DeviceRelationship.DeviceID.Eq(targetDeviceID)).
		Where(query.DeviceRelationship.TargetDeviceID.Eq(deviceID)).
		Where(query.DeviceRelationship.Status.Eq(model.RelationshipStatusAccepted.String())).
		First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		span.RecordError(err)
		return false, err
	}

	return rel2 != nil, nil
}

// CreateDeviceRelationship 创建设备关系，唯一索引冲突时更新 status
func (d *DeviceRelationshipRepo) CreateDeviceRelationship(ctx context.Context, deviceID, targetDeviceID string, status model.RelationshipStatus) error {
	_, span := tracer.Start(ctx, "DeviceRelationshipRepo.CreateDeviceRelationship")
	defer span.End()

	err := query.DeviceRelationship.WithContext(ctx).
		Clauses(
			clause.OnConflict{
				Columns: []clause.Column{
					{
						Name: query.DeviceRelationship.DeviceID.ColumnName().String(),
					},
					{
						Name: query.DeviceRelationship.TargetDeviceID.ColumnName().String(),
					},
				},
				DoUpdates: clause.AssignmentColumns([]string{
					query.DeviceRelationship.Status.ColumnName().String(),
					query.DeviceRelationship.UpdatedAt.ColumnName().String(),
				}),
			},
		).
		Create(&model.DeviceRelationship{
			DeviceID:       deviceID,
			TargetDeviceID: targetDeviceID,
			Status:         status,
			UpdatedAt:      time.Now(),
		})
	if err != nil {
		span.RecordError(err)
		return err
	}
	return nil
}

// DeleteDeviceRelationship 删除设备关系（双向删除：A->B 和 B->A）
func (d *DeviceRelationshipRepo) DeleteDeviceRelationship(ctx context.Context, deviceID, targetDeviceID string) error {
	_, span := tracer.Start(ctx, "DeviceRelationshipRepo.DeleteDeviceRelationship")
	defer span.End()

	// 删除双向关系：A->B 和 B->A
	_, err := query.DeviceRelationship.WithContext(ctx).
		Where(query.DeviceRelationship.DeviceID.Eq(deviceID), query.DeviceRelationship.TargetDeviceID.Eq(targetDeviceID)).
		Or(query.DeviceRelationship.DeviceID.Eq(targetDeviceID), query.DeviceRelationship.TargetDeviceID.Eq(deviceID)).
		Delete()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID), attribute.String("target_device_id", targetDeviceID))
		return err
	}

	return nil
}

// GetFriendsByDeviceID 获取所有的朋友，在指定时间范围内的朋友
func (d *DeviceRelationshipRepo) GetFriendsByDeviceID(ctx context.Context, deviceID string, startTime, endTime time.Time) ([]*model.DeviceRelationship, error) {
	_, span := tracer.Start(ctx, "DeviceRelationshipRepo.GetFriendsByDeviceID")
	defer span.End()

	var reverseDeviceIDs []string
	err := query.DeviceRelationship.WithContext(ctx).
		Where(query.DeviceRelationship.TargetDeviceID.Eq(deviceID)).
		Where(query.DeviceRelationship.Status.Eq(model.RelationshipStatusAccepted.String())).
		Pluck(query.DeviceRelationship.DeviceID, &reverseDeviceIDs)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	if len(reverseDeviceIDs) == 0 {
		return []*model.DeviceRelationship{}, nil
	}

	return query.DeviceRelationship.WithContext(ctx).
		Where(query.DeviceRelationship.DeviceID.Eq(deviceID)).
		Where(query.DeviceRelationship.Status.Eq(model.RelationshipStatusAccepted.String())).
		Where(query.DeviceRelationship.TargetDeviceID.In(reverseDeviceIDs...)).
		Where(query.DeviceRelationship.CreatedAt.Between(startTime, endTime)).
		Find()
}
