package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"context"
	"errors"
	"log/slog"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gen"
	"gorm.io/gen/field"
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
	_, span := tracer.Start(ctx, "FirstAddDevice")
	defer span.End()

	db := query.Q
	if len(tx) > 0 {
		db = tx[0]
	}
	slog.Info("[DeviceRepo] FirstAddDevice", "device_id", deviceID, "uid", uid)
	device, err := db.Device.Where(db.Device.DeviceID.Eq(deviceID)).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		span.SetAttributes(attribute.Int64("uid", uid))
		span.SetAttributes(attribute.String("error", err.Error()))
		slog.Error("[DeviceRepo] FirstAddDevice error", "device_id", deviceID, "uid", uid, "error", err.Error())
		return err
	}

	if device != nil && device.UID != uid {
		return errors.New("设备已绑定其他用户")
	}

	if device != nil {
		span.SetAttributes(attribute.String("device_id", deviceID))
		span.SetAttributes(attribute.Int64("uid", uid))
		slog.Info("[DeviceRepo] FirstAddDevice device already exists", "device_id", deviceID, "uid", uid)
		return nil
	}

	err = db.Device.Create(&model.Device{
		DeviceID: deviceID,
		UID:      uid,
		IsAdmin:  true,
	})
	if err != nil {
		slog.Error("[DeviceRepo] FirstAddDevice create error", "device_id", deviceID, "uid", uid, "error", err.Error())
	} else {
		slog.Info("[DeviceRepo] FirstAddDevice success", "device_id", deviceID, "uid", uid)
	}
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

// ChangeDeviceInfo 更换设备 SIM 卡号
func (d *DeviceRepo) ChangeDeviceInfo(ctx context.Context, deviceID string, simCard string, boardType, version string, relation string, tx ...*query.Query) error {
	_, span := tracer.Start(ctx, "ChangeDeviceInfo")
	defer span.End()

	db := query.Q
	if len(tx) > 0 {
		db = tx[0]
	}

	if _, err := db.Device.Where(db.Device.DeviceID.Eq(deviceID)).
		Updates(map[string]interface{}{
			db.Device.SIMCard.ColumnName().String():   simCard,
			db.Device.BoardType.ColumnName().String(): boardType,
			db.Device.Version.ColumnName().String():   version,
			db.Device.Relation.ColumnName().String():  relation,
		}); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		span.SetAttributes(attribute.String("sim_card", simCard))
		span.SetAttributes(attribute.String("board_type", boardType))
		span.SetAttributes(attribute.String("version", version))
		span.SetAttributes(attribute.String("relation", relation))
		span.SetAttributes(attribute.String("error", err.Error()))

		slog.Error("[DeviceRepo] ChangeDeviceInfo error", "device_id", deviceID, "sim_card", simCard, "board_type", boardType, "version", version, "relation", relation, "error", err.Error())
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
	// slog.Info("[DeviceRepo] CheckDeviceAuth", "uid", uid, "device_id", deviceID)
	// 同时匹配大小写，因为数据库中可能存储大写或小写
	lowerDeviceID := strings.ToLower(deviceID)
	upperDeviceID := strings.ToUpper(deviceID)
	num, err := query.Device.Debug().Where(
		query.Device.UID.Eq(uid),
		field.Or(
			query.Device.DeviceID.Eq(lowerDeviceID),
			query.Device.DeviceID.Eq(upperDeviceID),
		),
	).Count()
	if err != nil {
		span.RecordError(err)
		slog.Error("[DeviceRepo] CheckDeviceAuth error", "uid", uid, "device_id", deviceID, "error", err.Error())
		span.SetAttributes(attribute.Int64("uid", uid))
		span.SetAttributes(attribute.String("device_id", deviceID))
		span.SetAttributes(attribute.String("error", err.Error()))
		return false
	}
	// slog.Info("[DeviceRepo] CheckDeviceAuth num", "num", num)
	return num > 0
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

		// 纪念日
		_, err = tx.AnniversaryReminder.Where(tx.AnniversaryReminder.DeviceID.Eq(deviceID)).Delete()
		if err != nil {
			return err
		}

		// 消息
		_, err = tx.DeviceMessage.Where(tx.DeviceMessage.FromDeviceID.Eq(deviceID)).
			Or(tx.DeviceMessage.ToDeviceID.Eq(deviceID)).
			Delete()
		if err != nil {
			return err
		}

		// 事件提醒
		_, err = tx.Reminder.Where(tx.Reminder.DeviceID.Eq(deviceID)).
			Delete()
		if err != nil {
			return err
		}

		// 情绪
		_, err = tx.Emotion.Where(tx.Emotion.DeviceID.Eq(deviceID)).
			Delete()
		if err != nil {
			return err
		}

		// nfc
		_, err = tx.NFC.Where(tx.NFC.DeviceID.Eq(deviceID)).
			Delete()
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
func (d *DeviceRepo) FindUserInfoByDeviceID(ctx context.Context, deviceID string) (*model.User, error) {
	_, span := tracer.Start(ctx, "DeviceService.FindUserInfoByDeviceID")
	defer span.End()

	user, err := query.Device.Where(query.Device.DeviceID.Eq(deviceID)).
		Preload(query.Device.User).
		First()
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		return nil, err
	}
	if user == nil || user.User == nil {
		span.RecordError(errors.New("设备未绑定用户"))
		span.SetAttributes(attribute.String("device_id", deviceID))
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

// GetUserDeviceList 获取用户的设备列表
func (d *DeviceRepo) GetUserDeviceList(ctx context.Context, uid int64) ([]*model.Device, error) {
	_, span := tracer.Start(ctx, "DeviceService.GetUserDeviceList")
	defer span.End()

	devices, err := query.Device.Debug().
		Where(query.Device.UID.Eq(uid)).
		Preload(query.Device.DeviceInfo).
		Find()
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.Int64("uid", uid))
		slog.Error("[DeviceRepo] GetUserDeviceList", "uid", uid, "error", err.Error())
		return nil, err
	}
	return devices, nil
}

// SetDeviceAgent 设置设备Agent
func (d *DeviceRepo) SetDeviceAgent(deviceID string, agentName string) error {
	if _, err := query.Device.Where(query.Device.DeviceID.Eq(deviceID)).
		Updates(map[string]any{
			query.Device.AgentName.ColumnName().String(): agentName,
		}); err != nil {
		return err
	}
	return nil
}
