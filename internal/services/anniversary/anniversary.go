// Package anniversary 纪念日服务层
package anniversary

import (
	"aibuddy/cmd/server/task"
	"aibuddy/internal/model"
	"aibuddy/internal/repository"
	"aibuddy/pkg/config"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var tracer = func() trace.Tracer {
	return otel.Tracer(config.Instance.Tracer.ServiceName)
}

const (
	// TASKNAME 任务前缀
	TASKNAME = "anniversary"
)

// Service 服务层
type Service struct {
	DeviceRepo      *repository.DeviceRepo
	AnniversaryRepo *repository.AnniversaryRepo
	TaskClient      *task.Manager
}

// NewAnniversaryService 实例化服务层
func NewAnniversaryService() *Service {
	return &Service{
		DeviceRepo: repository.NewDeviceRepo(),
		TaskClient: task.NewManager(),
	}
}

// SubmitAnniversary 提交纪念日数据
func (s *Service) SubmitAnniversary(ctx context.Context, uid int64, deviceID string, data *model.AnniversaryReminder) error {
	ctx, span := tracer().Start(ctx, "AddRemind")
	defer span.End()

	if !s.DeviceRepo.CheckDeviceAuth(ctx, uid, deviceID) {
		return errors.New("无操作该设备信息的权限")
	}

	if err := s.AnniversaryRepo.Upsert(ctx, data); err != nil {
		return errors.New("提交纪念日数据失败:" + err.Error())
	}

	// 纪念日定时任务
	taskID := fmt.Sprintf("%s_%d", TASKNAME, data.ID)
	info, _ := s.TaskClient.GetTaskInfoByID("default", taskID)
	if info != nil {
		err := s.TaskClient.CancelTask("default", taskID)
		if err != nil {
			span.RecordError(err)
			return errors.New("添加提醒任务失败:%s" + err.Error())
		}
	}

	remindTime := s.CalRemindTime(data.AnniversaryTime, string(data.ReminderWay))
	payload := []byte(fmt.Sprintf(`{"scheduled":%v,"id":%d}`, true, data.ID))
	task := asynq.NewTask("anniversary", payload)
	info, err := s.TaskClient.EnqueueAt(task, remindTime, asynq.TaskID(taskID))
	if err != nil {
		span.RecordError(err)
		return errors.New("添加提醒任务失败:%s" + err.Error())
	}

	if _, err := s.AnniversaryRepo.UpdateEntryID(data.ID, info.ID); err != nil {
		return errors.New("添加提醒任务失败:%s" + err.Error())
	}

	return nil
}

// DeleteAnniversary 删除纪念日数据
func (s *Service) DeleteAnniversary(ctx context.Context, uid int64, deviceID string, id int64) error {
	ctx, span := tracer().Start(ctx, "AddRemind")
	defer span.End()
	if !s.DeviceRepo.CheckDeviceAuth(ctx, uid, deviceID) {
		return errors.New("无操作该设备信息的权限")
	}

	if err := s.AnniversaryRepo.Delete(ctx, id, deviceID); err != nil {
		return err
	}

	// 纪念日定时任务
	taskID := fmt.Sprintf("%s_%d", TASKNAME, id)
	info, _ := s.TaskClient.GetTaskInfoByID("default", taskID)
	if info != nil {
		return s.TaskClient.CancelTask("default", taskID)
	}

	return nil
}

// GetListByPage 获取分页数据
func (s *Service) GetListByPage(ctx context.Context, uid int64, deviceID string, page int, size int) ([]*model.AnniversaryReminder, int64, error) {
	ctx, span := tracer().Start(ctx, "AddRemind")
	defer span.End()
	if !s.DeviceRepo.CheckDeviceAuth(ctx, uid, deviceID) {
		return nil, 0, errors.New("无操作该设备信息的权限")
	}

	return s.AnniversaryRepo.GetListByPage(ctx, deviceID, page, size)
}

// CalRemindTime 计算提醒时间
func (s *Service) CalRemindTime(anniversaryDate time.Time, way string) time.Time {
	addTime := 0

	switch way {
	case string(model.ReminderWayOnDay):
		addTime = 0
	case string(model.ReminderWayBeforeOneDay):
		addTime = -1
	case string(model.ReminderWayBeforeThreeDays):
		addTime = -3
	case string(model.ReminderWayBeforeOneWeek):
		addTime = -7
	}

	remindTime := anniversaryDate.AddDate(0, 0, addTime)

	// 超出当前时间 需要下一年提醒
	if time.Now().Unix() > remindTime.Unix() {
		remindTime = remindTime.AddDate(1, 0, 0)
	}

	return remindTime
}
