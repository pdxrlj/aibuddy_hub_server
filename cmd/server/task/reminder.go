// Package task provides task management functionality
package task

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
)

// Task types
const (
	// 提醒任务
	TaskTypeReminder = "reminder"
)

// ReminderTaskPayload 提醒任务负载
type ReminderTaskPayload struct {
	RemindID int `json:"remind_id"`
}

// ReminderHandler 提醒任务处理函数
func ReminderHandler(_ context.Context, t *asynq.Task) error {
	var payload ReminderTaskPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	return nil
}
