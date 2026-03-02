package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"context"

	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm/clause"
)

// DeviceInfoRepo 设备信息仓库
type DeviceInfoRepo struct{}

// NewDeviceInfoRepo 实例化设备信息仓库
func NewDeviceInfoRepo() *DeviceInfoRepo {
	return &DeviceInfoRepo{}
}

// UpsertProfile 更新或者插入信息
func (d *DeviceInfoRepo) UpsertProfile(ctx context.Context, info *model.DeviceInfo) error {
	_, span := tracer.Start(ctx, "UpsertProfile")
	defer span.End()
	if err := query.DeviceInfo.Clauses(
		clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]any{
				query.DeviceInfo.NickName.ColumnName().String():    info.NickName,
				query.DeviceInfo.Avatar.ColumnName().String():      info.Avatar,
				query.DeviceInfo.Birthday.ColumnName().String():    info.Birthday,
				query.DeviceInfo.Gender.ColumnName().String():      info.Gender,
				query.DeviceInfo.Relation.ColumnName().String():    info.Relation,
				query.DeviceInfo.Hobbies.ColumnName().String():     info.Hobbies,
				query.DeviceInfo.Values.ColumnName().String():      info.Values,
				query.DeviceInfo.Skills.ColumnName().String():      info.Skills,
				query.DeviceInfo.Personality.ColumnName().String(): info.Personality,
			}),
		},
	).Create(info); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error", err.Error()))
		return err
	}
	return nil
}
