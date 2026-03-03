package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"context"

	"gorm.io/gorm/clause"
)

// DeviceOtaRepo 设备OTA仓库
type DeviceOtaRepo struct {
}

// NewDeviceOtaRepo 创建设备OTA仓库
func NewDeviceOtaRepo() *DeviceOtaRepo {
	return &DeviceOtaRepo{}
}

// UpsertDeviceOta 创建或更新设备OTA记录
func (d *DeviceOtaRepo) UpsertDeviceOta(ctx context.Context, deviceOta *model.DeviceOta) error {
	_, span := tracer.Start(ctx, "UpsertDeviceOta")
	defer span.End()

	if err := query.DeviceOta.Clauses(
		clause.OnConflict{
			Columns: []clause.Column{
				{Name: query.DeviceOta.DeviceID.ColumnName().String()},
			},
			DoUpdates: clause.AssignmentColumns([]string{
				query.DeviceOta.Version.ColumnName().String(),
				query.DeviceOta.OtaURL.ColumnName().String(),
				query.DeviceOta.ModelURL.ColumnName().String(),
				query.DeviceOta.ResourceURL.ColumnName().String(),
				query.DeviceOta.ForceUpdate.ColumnName().String(),
			}),
		},
	).WithContext(ctx).Create(deviceOta); err != nil {
		span.RecordError(err)
		return err
	}
	return nil
}
