package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"context"

	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gen"
	"gorm.io/gorm/clause"
)

// AnniversaryRepo 仓库
type AnniversaryRepo struct{}

// NewAnniversaryRepo 创建Anniversary仓库实例
func NewAnniversaryRepo() *AnniversaryRepo {
	return &AnniversaryRepo{}
}

// Upsert 更新/插入纪念日数据
func (a *AnniversaryRepo) Upsert(ctx context.Context, data *model.AnniversaryReminder) error {
	_, span := tracer.Start(ctx, "UpsertAnniversary")
	defer span.End()
	if err := query.AnniversaryReminder.Clauses(
		clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]any{
				query.AnniversaryReminder.DeviceID.ColumnName().String():         data.DeviceID,
				query.AnniversaryReminder.AnniversaryType.ColumnName().String():  data.AnniversaryType,
				query.AnniversaryReminder.ReminderUsername.ColumnName().String(): data.ReminderUsername,
				query.AnniversaryReminder.ReminderUserSex.ColumnName().String():  data.ReminderUserSex,
				query.AnniversaryReminder.AnniversaryTime.ColumnName().String():  data.AnniversaryTime,
				query.AnniversaryReminder.ReminderWay.ColumnName().String():      data.ReminderWay,
				query.AnniversaryReminder.Remarks.ColumnName().String():          data.Remarks,
			}),
		},
	).Create(data); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error", err.Error()))
		return err
	}

	return nil
}

// Delete 删除纪念日数据
func (a *AnniversaryRepo) Delete(ctx context.Context, id int64, deviceID string) error {
	_, span := tracer.Start(ctx, "DeleteAnniversary")
	defer span.End()
	_, err := query.AnniversaryReminder.
		Where(query.AnniversaryReminder.ID.Eq(id), query.AnniversaryReminder.DeviceID.Eq(deviceID)).
		Delete()

	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error", err.Error()))
		return err
	}
	return nil
}

// GetListByPage 获取分页数据
func (a *AnniversaryRepo) GetListByPage(ctx context.Context, deviceID string, page int, size int) ([]*model.AnniversaryReminder, int64, error) {
	_, span := tracer.Start(ctx, "GetListByPageAnniversary")
	defer span.End()
	offset := (page - 1) * size
	data, total, err := query.AnniversaryReminder.Where(query.AnniversaryReminder.DeviceID.Eq(deviceID)).FindByPage(offset, size)

	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, 0, err
	}
	return data, total, err
}

// UpdateEntryID 更新定时任务ID
func (a *AnniversaryRepo) UpdateEntryID(id int64, entryID string) (gen.ResultInfo, error) {
	return query.AnniversaryReminder.Where(query.AnniversaryReminder.ID.Eq(id)).Update(query.AnniversaryReminder.EntryID, entryID)
}

// GetAnniversaryByID 获取纪念日信息
func (a *AnniversaryRepo) GetAnniversaryByID(id int64) (*model.AnniversaryReminder, error) {
	return query.AnniversaryReminder.Where(query.AnniversaryReminder.ID.Eq(id)).First()
}
