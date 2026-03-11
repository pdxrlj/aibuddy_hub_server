// Package repository is the repository for the pomodoro clock
package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"context"
	"time"
)

// PomodoroClockRepo is the repository for the pomodoro clock
type PomodoroClockRepo struct {
}

// NewPomodoroClockRepo creates a new pomodoro clock repository
func NewPomodoroClockRepo() *PomodoroClockRepo {
	return &PomodoroClockRepo{}
}

// Create creates a new pomodoro clock
func (r *PomodoroClockRepo) Create(ctx context.Context, pomodoroClock *model.PomodoroClock) error {
	return query.PomodoroClock.WithContext(ctx).Create(pomodoroClock)
}

// GetPomodoroClockByDeviceID 获取设备在指定时间范围内的番茄钟列表
func (r *PomodoroClockRepo) GetPomodoroClockByDeviceID(ctx context.Context, deviceID string, startTime, endTime time.Time) ([]*model.PomodoroClock, error) {
	return query.PomodoroClock.WithContext(ctx).
		Where(query.PomodoroClock.DeviceID.Eq(deviceID)).
		Where(query.PomodoroClock.CreatedAt.Between(startTime, endTime)).
		Find()
}
