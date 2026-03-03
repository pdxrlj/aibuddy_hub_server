package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"context"

	"gorm.io/gorm"
)

// DeviceRepo 设备信息仓库
type DeviceRepo struct{}

// NewDeviceRepo 实例化设备信息仓库
func NewDeviceRepo() *DeviceInfoRepo {
	return &DeviceInfoRepo{}
}

// FirstAddDevice 第一次绑定设备
func (d *DeviceRepo) FirstAddDevice(ctx context.Context, deviceID string, uid int64, tx ...*query.Query) error {
	_, span := tracer.Start(ctx, "UpsertProfile")
	defer span.End()

	db := query.Q
	if len(tx) > 0 {
		db = tx[0]
	}

	data, err := db.Device.Where(db.Device.DeviceID.Eq(deviceID), db.Device.UID.Eq(uid)).First()
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if data != nil {
		return nil
	}

	err = db.Device.Create(&model.Device{
		DeviceID: deviceID,
		UID:      uid,
		IsAdmin:  true,
	})
	return err
}

// ChangeDeviceRole 设备切换角色
func (d *DeviceRepo) ChangeDeviceRole(ctx context.Context, uid int64, deviceID string, roleID int64) error {
	_, span := tracer.Start(ctx, "ChangeDeviceRole")
	defer span.End()
	if _, err := query.Device.Where(query.Device.DeviceID.Eq(deviceID), query.Device.UID.Eq(uid), query.Device.IsAdmin.Is(true)).
		Update(query.Device.AgentID, roleID); err != nil {
		return err
	}
	return nil
}

// ChangeDeviceInfo 更换设备 ICCID
func (d *DeviceRepo) ChangeDeviceInfo(ctx context.Context, deviceID string, iccid string, boardType, version string, tx ...*query.Query) error {
	_, span := tracer.Start(ctx, "ChangeDeviceIccid")
	defer span.End()

	db := query.Q
	if len(tx) > 0 {
		db = tx[0]
	}

	if _, err := db.Device.Where(db.Device.DeviceID.Eq(deviceID)).
		Updates(map[string]interface{}{
			db.Device.ICCID.ColumnName().String():     iccid,
			db.Device.BoardType.ColumnName().String(): boardType,
			db.Device.Version.ColumnName().String():   version,
		}); err != nil {
		return err
	}

	return nil
}
