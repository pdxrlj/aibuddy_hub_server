package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gen"
	"gorm.io/gorm/clause"
)

// RemindRepo 提醒事件仓库
type RemindRepo struct{}

// NewRemindRepo 实例化提醒事件仓库
func NewRemindRepo() *RemindRepo {
	return &RemindRepo{}
}

// InsertOrUpdate 更新/插入数据
func (r *RemindRepo) InsertOrUpdate(ctx context.Context, data *model.Reminder) error {
	_, span := tracer.Start(ctx, "InsertOrUpdate")
	defer span.End()

	if data.ReminderTime.Unix() < time.Now().Unix() {
		return errors.New("提醒时间不能是过去时间")
	}

	if !model.IsValidRepeatType(data.RepeatType) {
		return errors.New("不存在的重复类型")
	}
	if !model.IsValidReminderStatus(data.Status) {
		return errors.New("不存在的重复类型")
	}
	if data.ID > 0 {
		count, err := query.Reminder.Where(query.Reminder.ID.Eq(data.ID)).Count()
		if err != nil || count < 1 {
			return errors.New("不存在的事件ID")
		}
	}

	if err := query.Reminder.Clauses(
		clause.OnConflict{
			Columns: []clause.Column{{Name: query.Reminder.ID.ColumnName().String()}},
			DoUpdates: clause.Assignments(map[string]any{
				query.Reminder.RepeatType.ColumnName().String():      data.RepeatType,
				query.Reminder.ReminderTitle.ColumnName().String():   data.ReminderTitle,
				query.Reminder.ReminderContent.ColumnName().String(): data.ReminderContent,
				query.Reminder.ReminderTime.ColumnName().String():    data.ReminderTime,
				query.Reminder.DeviceID.ColumnName().String():        data.DeviceID,
				query.Reminder.Status.ColumnName().String():          data.Status,
			}),
		},
	).Create(data); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error", err.Error()))
		return err
	}

	return nil
}

// GetDeviceID 获取用户device_id
func (r *RemindRepo) GetDeviceID(ctx context.Context, id int64) (string, error) {
	_, span := tracer.Start(ctx, "GetDeviceID")
	defer span.End()

	data, err := query.Reminder.Select(query.Reminder.DeviceID).Where(query.Reminder.ID.Eq(id)).First()
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error", err.Error()))
		return "", err
	}

	return data.DeviceID, nil
}

// DeleteByID 删除提醒事件
func (r *RemindRepo) DeleteByID(ctx context.Context, id int64) error {
	_, span := tracer.Start(ctx, "DeleteByID")
	defer span.End()

	_, err := query.Reminder.Where(query.Reminder.ID.Eq(id)).Delete()
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error", err.Error()))
		return err
	}

	return nil
}

// GetListByDeviceID 获取指定设备的提醒事件列表
func (r *RemindRepo) GetListByDeviceID(ctx context.Context, deviceID string, page int64, size int64) ([]*model.Reminder, int64, error) {
	_, span := tracer.Start(ctx, "DeleteByID")
	defer span.End()
	offset := (page - 1) * size
	return query.Reminder.Where(query.Reminder.DeviceID.Eq(deviceID)).FindByPage(int(offset), int(size))
}

// UpdateEntryID 更新定时任务ID
func (r *RemindRepo) UpdateEntryID(id int64, entryID string) (gen.ResultInfo, error) {
	return query.Reminder.Where(query.Reminder.ID.Eq(id)).Update(query.Reminder.EntryID, entryID)
}
