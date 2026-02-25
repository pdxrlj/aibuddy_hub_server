// Package schedule provides task scheduling functionality.
package schedule

import (
	"aibuddy/pkg/config"
	"aibuddy/pkg/schedule"
	"context"
	"log/slog"
)

// StartSchedule starts the task scheduler.
func StartSchedule(ctx context.Context) error {
	scheduleConfig := &schedule.Config{
		RedisDB:       config.Instance.Storage.Redis.DB,
		RedisHost:     config.Instance.Storage.Redis.Host,
		RedisPort:     config.Instance.Storage.Redis.Port,
		RedisPassword: config.Instance.Storage.Redis.Password,
	}
	schedule, err := schedule.New(scheduleConfig)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		if err := schedule.Shutdown(); err != nil {
			slog.Error("Failed to shutdown schedule", "error", err)
		}
	}()

	if err := RegisterTasks(); err != nil {
		return err
	}

	return schedule.Consumer()
}
