// Package remind  remind 提醒事件服务
package remind

import (
	"aibuddy/internal/model"
	"aibuddy/internal/repository"
	"aibuddy/pkg/config"
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var tracer = func() trace.Tracer {
	return otel.Tracer(config.Instance.Tracer.ServiceName)
}

// Service 提醒事件服务
type Service struct {
	DeviceRepo *repository.DeviceRepo
	RemindRepo *repository.RemindRepo
}

// NewRemindService 实例化提醒事件服务
func NewRemindService() *Service {
	return &Service{
		DeviceRepo: repository.NewDeviceRepo(),
		RemindRepo: repository.NewRemindRepo(),
	}
}

// SubmitRemind 添加提醒事件
func (r *Service) SubmitRemind(ctx context.Context, uid int64, data *model.Reminder) error {
	ctx, span := tracer().Start(ctx, "AddRemind")
	defer span.End()

	if !r.DeviceRepo.CheckDeviceAuth(ctx, uid, data.DeviceID) {
		return errors.New("无操作该设备信息的权限")
	}

	if err := r.RemindRepo.InsertOrUpdate(ctx, data); err != nil {
		return err
	}

	return nil
}

// DeleateRemindByID 删除提醒事件
func (r *Service) DeleateRemindByID(ctx context.Context, uid int64, id int64) error {
	ctx, span := tracer().Start(ctx, "DeleateRemindByID")
	defer span.End()

	deviceID, err := r.RemindRepo.GetDeviceID(ctx, id)
	if err != nil {
		return errors.New("无法删除提醒事件")
	}

	if !r.DeviceRepo.CheckDeviceAuth(ctx, uid, deviceID) {
		return errors.New("无操作权限")
	}

	return r.RemindRepo.DeleteByID(ctx, id)
}

// GetList 获取提醒事件列表
func (r *Service) GetList(ctx context.Context, uid int64, deviceID string, page int64, size int64) ([]*model.Reminder, int64, error) {
	ctx, span := tracer().Start(ctx, "GetRemindList")
	defer span.End()

	if !r.DeviceRepo.CheckDeviceAuth(ctx, uid, deviceID) {
		return nil, 0, errors.New("无操作权限")
	}

	return r.RemindRepo.GetListByDeviceID(ctx, deviceID, page, size)
}
