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
	"math/rand/v2"
	"time"

	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

// Task types
const (
	// 纪念日任务
	TaskTypeAnniversary = "anniversary"
)

// AnniversaryTaskPayload 提醒任务负载
type AnniversaryTaskPayload struct {
	RemindID  int64 `json:"remind_id"`
	Scheduled bool  `json:"scheduled"`
}

// AnniversaryHandler 提醒任务处理函数
func AnniversaryHandler(_ context.Context, t *asynq.Task) error {
	a := repository.NewAnniversaryRepo()
	var payload AnniversaryTaskPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	info, err := a.GetAnniversaryByID(payload.RemindID)
	if err != nil {
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}

	// 发送至设备
	if err := remind.SendMessage(remind.MsgTypeRemind, info.EntryID, "text", string(info.AnniversaryType), info.Remarks, info.DeviceID, ""); err != nil {
		return err
	}

	if !payload.Scheduled {
		return nil
	}

	// 计算下个提示时间重新入队列
	nextReminderTime := calNextRemindTime(info.AnniversaryTime, info.ReminderWay.String())
	manager := NewManager()
	taskID := fmt.Sprintf("remind_%d_%d", info.ID, rand.IntN(100))
	newTask := asynq.NewTask("anniversary", t.Payload())
	result, err := manager.EnqueueAt(newTask, nextReminderTime, asynq.TaskID(taskID))
	if err != nil {
		return err
	}

	_, err = a.UpdateEntryID(info.ID, result.ID)
	if err != nil {
		return err
	}
	return nil
}

// calNextRemindTime 计算下个提示时间
func calNextRemindTime(remindTime time.Time, way string) time.Time {
	addYear := time.Now().Year() - remindTime.Year() + 1
	addDay := 0
	switch way {
	case model.ReminderWayOnDay.String():
		addDay = 0
	case model.ReminderWayBeforeOneDay.String():
		addDay = 1
	case model.ReminderWayBeforeThreeDays.String():
		addDay = 3
	case model.ReminderWayBeforeOneWeek.String():
		addDay = 7
	}
	nextRemindTime := remindTime.AddDate(addYear, 0, -addDay)
	return nextRemindTime
}
