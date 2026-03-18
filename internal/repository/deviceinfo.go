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
func (d *DeviceInfoRepo) UpsertProfile(ctx context.Context, info *model.DeviceInfo, tx ...*query.Query) error {
	_, span := tracer.Start(ctx, "UpsertProfile")
	defer span.End()
	db := query.Q
	if len(tx) > 0 {
		db = tx[0]
	}

	if err := db.DeviceInfo.Clauses(
		clause.OnConflict{
			Columns: []clause.Column{{Name: query.DeviceInfo.DeviceID.ColumnName().String()}},
			DoUpdates: clause.Assignments(map[string]any{
				db.DeviceInfo.NickName.ColumnName().String():    info.NickName,
				db.DeviceInfo.Avatar.ColumnName().String():      info.Avatar,
				db.DeviceInfo.Birthday.ColumnName().String():    info.Birthday,
				db.DeviceInfo.Gender.ColumnName().String():      info.Gender,
				db.DeviceInfo.Hobbies.ColumnName().String():     info.Hobbies,
				db.DeviceInfo.Values.ColumnName().String():      info.Values,
				db.DeviceInfo.Skills.ColumnName().String():      info.Skills,
				db.DeviceInfo.Personality.ColumnName().String(): info.Personality,
			}),
		},
	).Create(info); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error", err.Error()))
		return err
	}
	return nil
}

// GetUserInfoByDeivceID 获取用户设备信息
func (d *DeviceInfoRepo) GetUserInfoByDeivceID(ctx context.Context, deivceID string) (*model.DeviceInfo, error) {
	ctx, span := tracer.Start(ctx, "GetUserInfoByDeivceID")
	defer span.End()
	return query.DeviceInfo.WithContext(ctx).Where(query.DeviceInfo.DeviceID.Eq(deivceID)).First()
}

// UpdateDeviceInfo  修改设备信息
func (d *DeviceInfoRepo) UpdateDeviceInfo(ctx context.Context, data *model.DeviceInfo, uid int64, relation string) error {
	_, span := tracer.Start(ctx, "UpdateDeviceInfo")
	defer span.End()
	return query.Q.Transaction(func(tx *query.Query) error {
		_, err := tx.DeviceInfo.Where(query.DeviceInfo.DeviceID.Eq(data.DeviceID)).Updates(data)
		if err != nil {
			return err
		}

		_, err = tx.Device.Where(tx.Device.UID.Eq(uid), tx.Device.DeviceID.Eq(data.DeviceID), tx.Device.IsAdmin.Eq(true)).
			Update(tx.Device.Relation, relation)
		if err != nil {
			return err
		}

		return nil
	})
}
