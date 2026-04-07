package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

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

// SetDeviceVersion 设置设备版本
func (d *DeviceRepo) SetDeviceVersion(deviceID string, version string) (info gen.ResultInfo, err error) {
	return query.Device.Where(query.Device.DeviceID.Eq(deviceID)).Updates(map[string]any{
		query.Device.Version.ColumnName().String(): version,
	})
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

		_, err = tx.DeviceRelationship.
			Where(tx.DeviceRelationship.DeviceID.Eq(deviceID)).
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

// GetUserDeviceIsAdmin 获取用户绑定设备
func (d *DeviceRepo) GetUserDeviceIsAdmin(ctx context.Context, uid int64) ([]*model.Device, error) {
	_, span := tracer.Start(ctx, "DeviceService.GetUserDeviceList")
	defer span.End()

	devices, err := query.Device.
		Where(query.Device.UID.Eq(uid), query.Device.IsAdmin.Eq(true)).
		Find()
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.Int64("uid", uid))
		slog.Error("[DeviceRepo] GetUserDeviceList", "uid", uid, "error", err.Error())
		return nil, err
	}
	return devices, nil
}

// ClearDeviceInfo 清除用户数据
func (d *DeviceRepo) ClearDeviceInfo(ctx context.Context, uid int64, deviceIDs []string) error {
	ctx, span := tracer.Start(ctx, "DeviceService.ClearDeviceInfo")
	defer span.End()

	return query.Q.Transaction(func(tx *query.Query) error {
		// 1. 定义一个统一的清理任务类型
		type cleanupTask func() error

		// 2. 将所有操作包装成统一的匿名函数
		// 这样可以屏蔽不同 table 之间 Delete() 返回值和参数的微小差异
		tasks := []cleanupTask{
			func() error {
				_, err := tx.AnniversaryReminder.WithContext(ctx).Where(tx.AnniversaryReminder.DeviceID.In(deviceIDs...)).Delete()
				return err
			},
			func() error {
				_, err := tx.ChatDialogue.WithContext(ctx).Where(tx.ChatDialogue.DeviceID.In(deviceIDs...)).Delete()
				return err
			},
			func() error {
				_, err := tx.Device.WithContext(ctx).Where(tx.Device.DeviceID.In(deviceIDs...)).Delete()
				return err
			},
			func() error {
				_, err := tx.DeviceInfo.WithContext(ctx).Where(tx.DeviceInfo.DeviceID.In(deviceIDs...)).Delete()
				return err
			},
			func() error {
				_, err := tx.DeviceOta.WithContext(ctx).Where(tx.DeviceOta.DeviceID.In(deviceIDs...)).Delete()
				return err
			},
			func() error {
				_, err := tx.Emotion.WithContext(ctx).Where(tx.Emotion.DeviceID.In(deviceIDs...)).Delete()
				return err
			},
			func() error {
				_, err := tx.NFC.WithContext(ctx).Where(tx.NFC.DeviceID.In(deviceIDs...)).Delete()
				return err
			},
			func() error {
				_, err := tx.PomodoroClock.WithContext(ctx).Where(tx.PomodoroClock.DeviceID.In(deviceIDs...)).Delete()
				return err
			},
			func() error {
				_, err := tx.Reminder.WithContext(ctx).Where(tx.Reminder.DeviceID.In(deviceIDs...)).Delete()
				return err
			},
			func() error {
				_, err := tx.UserAgent.WithContext(ctx).Where(tx.UserAgent.DeviceID.In(deviceIDs...)).Delete()
				return err
			},

			// 逻辑稍微复杂的也可以塞进来
			func() error {
				_, err := tx.DeviceMessage.WithContext(ctx).Where(
					tx.DeviceMessage.ToDeviceID.In(deviceIDs...),
				).Or(tx.DeviceMessage.FromDeviceID.In(deviceIDs...)).Delete()
				return err
			},
			func() error {
				_, err := tx.DeviceRelationship.WithContext(ctx).Where(
					tx.DeviceRelationship.DeviceID.In(deviceIDs...),
				).Or(tx.DeviceRelationship.TargetDeviceID.In(deviceIDs...)).Delete()
				return err
			},

			// 用户相关数据清理
			func() error { return d.ClearUserInfo(ctx, uid) },
		}

		// 3. 循环执行。现在整个 Transaction 闭包的圈复杂度只有 2（一个循环判断）
		for _, task := range tasks {
			if err := task(); err != nil {
				return err
			}
		}

		return nil
	})
}

// ClearUserInfo 清除用户
func (d *DeviceRepo) ClearUserInfo(ctx context.Context, uid int64) error {
	ctx, span := tracer.Start(ctx, "DeviceService.ClearUserInfo")
	defer span.End()
	return query.Q.Transaction(func(tx *query.Query) error {
		// 删除user
		_, err := tx.User.WithContext(ctx).Where(tx.User.ID.Eq(uid)).Delete()
		if err != nil {
			return err
		}

		// 删除用户反馈表
		_, err = tx.Feedback.WithContext(ctx).Where(tx.Feedback.UID.Eq(uid)).Delete()
		if err != nil {
			return err
		}

		return nil
	})
}

// SetDeviceAgent 设置设备Agent
func (d *DeviceRepo) SetDeviceAgent(deviceID string, agentName string, tx ...*query.Query) error {
	db := query.Q
	if len(tx) > 0 {
		db = tx[0]
	}

	_, err := db.Device.Where(db.Device.DeviceID.Eq(deviceID)).
		Update(db.Device.AgentName, agentName)
	if err != nil {
		slog.Error("[SetDeviceAgent] Update error", "error", err, "device_id", deviceID)
		return err
	}
	return nil
}

// SetDeviceVipExpireTime 设置设备VIP的过期时间（累加时长）
func (d *DeviceRepo) SetDeviceVipExpireTime(ctx context.Context, deviceID string, duration time.Duration, tx ...*query.Query) error {
	_, span := tracer.Start(ctx, "DeviceService.SetDeviceVipExpireTime")
	defer span.End()

	db := query.Q
	if len(tx) > 0 {
		db = tx[0]
	}

	// 查询设备当前的过期时间
	device, err := db.Device.Where(db.Device.DeviceID.Eq(deviceID)).First()
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		slog.Error("[SetDeviceVipExpireTime] Get device error", "error", err, "device_id", deviceID)
		return err
	}

	// 计算新的过期时间
	var newExpireTime time.Time
	now := time.Now()

	// 如果当前有过期时间且未过期，在当前基础上增加
	if !device.ExpireTime.IsZero() && device.ExpireTime.After(now) {
		newExpireTime = device.ExpireTime.Add(duration)
	} else {
		// 否则从当前时间开始计算
		newExpireTime = now.Add(duration)
	}

	// 更新过期时间
	_, err = db.Device.Where(db.Device.DeviceID.Eq(deviceID)).
		Update(db.Device.ExpireTime, newExpireTime)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(
			attribute.String("device_id", deviceID),
			attribute.String("new_expire_time", newExpireTime.Format("2006-01-02 15:04:05")),
		)
		slog.Error("[SetDeviceVipExpireTime] Update error", "error", err, "device_id", deviceID)
		return err
	}

	slog.Info("[SetDeviceVipExpireTime] Success",
		"device_id", deviceID,
		"duration", duration.String(),
		"new_expire_time", newExpireTime.Format("2006-01-02 15:04:05"))

	return nil
}

// IsDeviceVipExpired 判断设备VIP是否已经过期了
func (d *DeviceRepo) IsDeviceVipExpired(ctx context.Context, deviceID string) (bool, error) {
	_, span := tracer.Start(ctx, "DeviceService.IsDeviceVipExpired")
	defer span.End()

	deviceInfo, err := query.Device.Where(query.Device.DeviceID.Eq(deviceID)).First()
	if err != nil || deviceInfo == nil {
		return true, errors.New("没有查询到当前设备")
	}

	// 如果过期时间为空，则认为没有过期
	if deviceInfo.ExpireTime.IsZero() {
		return false, nil
	}

	// 判断过期时间是否小于当前时间
	return deviceInfo.ExpireTime.Before(time.Now()), nil
}

// SetVoiceID 设置设备的语音ID
func (d *DeviceRepo) SetVoiceID(ctx context.Context, deviceID string, voiceID string) error {
	_, span := tracer.Start(ctx, "DeviceService.SetVoiceID")
	defer span.End()

	_, err := query.Device.
		Where(query.Device.DeviceID.Eq(deviceID)).
		Update(query.Device.VoiceID, voiceID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		span.SetAttributes(attribute.String("voice_id", voiceID))
		slog.Error("[SetVoiceId] Update error", "error", err, "device_id", deviceID)
		return err
	}
	return nil
}
