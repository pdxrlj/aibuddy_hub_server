package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"context"
	"errors"

	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gen"
	"gorm.io/gorm"
)

// DeviceRepo 设备信息仓库
type DeviceRepo struct{}

// NewDeviceRepo 实例化设备信息仓库
func NewDeviceRepo() *DeviceRepo {
	return &DeviceRepo{}
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
func (d *DeviceRepo) ChangeDeviceRole(ctx context.Context, uid int64, deviceID string, roleName string) error {
	_, span := tracer.Start(ctx, "ChangeDeviceRole")
	defer span.End()
	if _, err := query.Device.Where(query.Device.DeviceID.Eq(deviceID), query.Device.UID.Eq(uid), query.Device.IsAdmin.Is(true)).
		Update(query.Device.AgentName, roleName); err != nil {
		return err
	}
	return nil
}

// ChangeDeviceInfo 更换设备 ICCID
func (d *DeviceRepo) ChangeDeviceInfo(ctx context.Context, deviceID string, iccid string, boardType, version string, relation string, tx ...*query.Query) error {
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
			db.Device.Relation.ColumnName().String():  relation,
		}); err != nil {
		return err
	}

	return nil
}

// BatchHandlerDeviceList 批量处理设备列表
func (d *DeviceRepo) BatchHandlerDeviceList(ctx context.Context, handeler func(devices []*model.Device)) error {
	_, span := tracer.Start(ctx, "GetDeviceList")
	defer span.End()

	var devices []*model.Device
	err := query.Device.FindInBatches(&devices, 100, func(_ gen.Dao, _ int) error {
		handeler(devices)
		return nil
	})
	if err != nil {
		span.RecordError(err)
		return err
	}
	return nil
}

// UpdateDeviceHardwareInfo 更新设备硬件信息
func (d *DeviceRepo) UpdateDeviceHardwareInfo(deviceID string, hardwareInfo []byte) error {
	if _, err := query.Device.Where(query.Device.DeviceID.Eq(deviceID)).
		Updates(map[string]any{
			query.Device.HardwareInfo.ColumnName().String(): hardwareInfo,
		}); err != nil {
		return err
	}
	return nil
}

// SetDeviceStatus 设置设备状态
func (d *DeviceRepo) SetDeviceStatus(deviceID string, state model.DeviceStatus) error {
	if _, err := query.Device.Where(query.Device.DeviceID.Eq(deviceID)).
		Updates(map[string]any{
			query.Device.Status.ColumnName().String(): state,
		}); err != nil {
		return err
	}
	return nil
}

// CheckDeviceAuth 是否拥有操作设备权限
func (d *DeviceRepo) CheckDeviceAuth(ctx context.Context, uid int64, deviceID string) bool {
	_, span := tracer.Start(ctx, "CheckDeviceAuth")
	defer span.End()

	num, err := query.Device.Where(query.Device.UID.Eq(uid), query.Device.DeviceID.Eq(deviceID)).Count()
	if err != nil {
		span.RecordError(err)
		return false
	}

	if num > 0 {
		return true
	}

	return false
}

// IsValidDevice 判断设备是否是有效的设备
func (d *DeviceRepo) IsValidDevice(ctx context.Context, deviceID string) bool {
	_, span := tracer.Start(ctx, "IsValidDevice")
	defer span.End()

	num, err := query.Device.Where(query.Device.DeviceID.Eq(deviceID)).Count()
	if err != nil {
		span.RecordError(err)
		return false
	}

	if num > 0 {
		return true
	}

	return false
}

// EraseDevice 擦除设备记录-device表，device_info表，device_relationship表
func (d *DeviceRepo) EraseDevice(ctx context.Context, deviceID string) error {
	_, span := tracer.Start(ctx, "EraseDevice")
	defer span.End()

	return query.Q.Transaction(func(tx *query.Query) error {
		_, err := tx.Device.Where(tx.Device.DeviceID.Eq(deviceID)).Delete()
		if err != nil {
			return err
		}

		_, err = tx.DeviceInfo.Where(tx.DeviceInfo.DeviceID.Eq(deviceID)).Delete()
		if err != nil {
			return err
		}

		_, err = tx.DeviceRelationship.Where(tx.DeviceRelationship.DeviceID.Eq(deviceID)).
			Or(query.DeviceRelationship.TargetDeviceID.Eq(deviceID)).
			Delete()
		if err != nil {
			return err
		}
		return nil
	})
}

// FindUserInfoByDeviceID 根据设备ID查询用户信息
func (d *DeviceRepo) FindUserInfoByDeviceID(deviceID string) (*model.User, error) {
	user, err := query.Device.Where(query.Device.DeviceID.Eq(deviceID)).
		Preload(query.Device.User).
		First()
	if err != nil {
		return nil, err
	}
	if user == nil || user.User == nil {
		return nil, errors.New("设备未绑定用户")
	}

	return user.User, nil
}

// GetDeviceInfo 获取设备信息
func (d *DeviceRepo) GetDeviceInfo(ctx context.Context, deviceID string) (*model.Device, error) {
	_, span := tracer.Start(ctx, "DeviceService.GetDeviceInfo")
	defer span.End()

	device, err := query.Device.
		Preload(query.Device.DeviceInfo).
		Where(query.Device.DeviceID.Eq(deviceID)).
		First()
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		return nil, err
	}
	return device, nil
}
