// Package remind  remind 提醒事件服务
package remind

import (
	"aibuddy/cmd/server/task"
	"aibuddy/internal/model"
	"aibuddy/internal/repository"
	"aibuddy/pkg/config"
	"context"
	"errors"
	"fmt"

	"github.com/hibiken/asynq"
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
	TaskClient *task.Manager
}

// NewRemindService 实例化提醒事件服务
func NewRemindService() *Service {
	return &Service{
		DeviceRepo: repository.NewDeviceRepo(),
		RemindRepo: repository.NewRemindRepo(),
		TaskClient: task.NewManager(),
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

	// 事件提醒-定时任务处理
	taskID := fmt.Sprintf("remind-%d", data.ID)
	if data.Status != model.ReminderStatusCancelled {
		scheduled := true
		if data.RepeatType == model.RepeatTypeNone {
			scheduled = false
		}
		payload := []byte(fmt.Sprintf(`{"scheduled":%v,"id":%d}`, scheduled, data.ID))
		task := asynq.NewTask("scheduled_task", payload)
		info, err := r.TaskClient.EnqueueAt(task, data.ReminderTime, asynq.TaskID(taskID))
		if err != nil {
			span.RecordError(err)
			return errors.New("添加提醒任务失败:%s" + err.Error())
		}

		fmt.Printf("%+v\n", info)
	} else {
		info, _ := r.TaskClient.GetTaskInfoByID("default", taskID)
		if info != nil {
			return r.TaskClient.CancelTask("default", taskID)
		}
	}
	return nil
}

// DeleateRemindByID 删除提醒事件
func (r *Service) DeleateRemindByID(ctx context.Context, uid int64, id int64) error {
	ctx, span := tracer().Start(ctx, "DeleateRemindByID")
	defer span.End()

	deviceID, err := r.RemindRepo.GetDeviceID(ctx, id)
	if err != nil {
		return errors.New("无法删除提醒事件或者提醒事件不存在")
	}

	if !r.DeviceRepo.CheckDeviceAuth(ctx, uid, deviceID) {
		return errors.New("无操作权限")
	}

	if err := r.RemindRepo.DeleteByID(ctx, id); err != nil {
		return errors.New("删除提醒事件失败:" + err.Error())
	}

	// 事件提醒-定时任务处理
	taskID := fmt.Sprintf("remind-%d", id)
	info, _ := r.TaskClient.GetTaskInfoByID("default", taskID)
	if info != nil {
		return r.TaskClient.CancelTask("default", taskID)
	}

	return nil
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
