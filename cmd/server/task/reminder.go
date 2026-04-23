// Package task provides task management functionality
package task

import (
	"aibuddy/aiframe/remind"
	"aibuddy/internal/model"
	"aibuddy/internal/repository"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

// Task types
const (
	// 提醒任务
	TaskTypeReminder = "reminder"
)

// ReminderTaskPayload 提醒任务负载
type ReminderTaskPayload struct {
	RemindID  int  `json:"remind_id"`
	Scheduled bool `json:"scheduled"`
}

// ReminderHandler 提醒任务处理函数
func ReminderHandler(_ context.Context, t *asynq.Task) error {
	r := repository.NewRemindRepo()
	var payload ReminderTaskPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	info, err := r.GetRemindByID(int64(payload.RemindID))
	if err != nil {
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}

	// 是否取消任务
	if info.Status == model.ReminderStatusCancelled {
		return nil
	}

	// 发送至设备
	if err := remind.SendMessage(remind.MsgTypeRemind, info.EntryID, "text", info.ReminderTitle, info.ReminderContent, info.DeviceID, ""); err != nil {
		return err
	}
	slog.Info("[Reminder] ReminderHandler")

	if !payload.Scheduled {
		if _, err := r.UpdateStatus(info.ID, model.ReminderStatusCompleted.String()); err != nil {
			return err
		}
		return nil
	}
	// 计算下次提醒事件
	info.NextReminderTime = cmputedNextReminderTime(info.ReminderTime, info.NextReminderTime, info.RepeatType)

	manager := NewManager()
	taskID := fmt.Sprintf("remind_%d_%d", info.ID, rand.IntN(100))
	newTask := asynq.NewTask("reminder", t.Payload())
	result, err := manager.EnqueueAt(newTask, info.NextReminderTime, asynq.TaskID(taskID))
	if err != nil {
		return err
	}

	_, err = r.UpdateRemindTask(info.ID, result.ID, info.NextReminderTime)
	if err != nil {
		return err
	}
	return nil
}

// cmputedNextReminderTime 计算下次提醒时间
func cmputedNextReminderTime(remindTime time.Time, nextTime time.Time, repeatType model.RepeatType) time.Time {
	if nextTime.Unix() <= 0 {
		nextTime = remindTime
	}
	switch repeatType {
	case model.RepeatTypeNone:
		nextTime = remindTime
	case model.RepeatTypeDaily:
		nextTime = remindTime.AddDate(0, 0, 1)
	case model.RepeatTypeWeekly:
		nextTime = remindTime.AddDate(0, 0, 7)
	case model.RepeatTypeMonthly:
		nextTime = remindTime.AddDate(0, 1, 0)
	case model.RepeatTypeYearly:
		nextTime = remindTime.AddDate(1, 0, 0)
	default:
		nextTime = remindTime
	}
	return nextTime
}
