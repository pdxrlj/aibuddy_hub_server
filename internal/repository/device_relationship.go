// Package repository provides a repository for the device relationship.
package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"context"

	"gorm.io/gen/field"
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
