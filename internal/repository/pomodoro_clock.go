// Package repository is the repository for the pomodoro clock
package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"context"
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
