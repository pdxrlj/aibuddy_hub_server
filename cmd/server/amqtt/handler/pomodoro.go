// Package handler is the pomodoro service
package handler

import (
	"aibuddy/aiframe/pomodoro"
	"aibuddy/internal/model"
	"aibuddy/internal/repository"
	"aibuddy/pkg/mqtt"
	"context"
	"log/slog"
)

// PomodoroHandler is the pomodoro handler
type PomodoroHandler struct {
	PomodoroClockRepo *repository.PomodoroClockRepo
}

// NewPomodoroHandler creates a new pomodoro handler
func NewPomodoroHandler() *PomodoroHandler {
	return &PomodoroHandler{
		PomodoroClockRepo: repository.NewPomodoroClockRepo(),
	}
}

// Handle handles the pomodoro message
func (h *PomodoroHandler) Handle(ctx *mqtt.Context) {
	defer ctx.Message.Ack()

	deviceID := ctx.Params["device_id"]
	var pomodoro pomodoro.Pomodoro
	if err := pomodoro.Decode(ctx.Payload); err != nil {
		slog.Error("[MQTT] PomodoroHandler decode failed", "device_id", deviceID, "error", err)
		return
	}

	pomodoroClock := &model.PomodoroClock{
		DeviceID:            deviceID,
		TotalDuration:       pomodoro.TotalTime,
		StudyDuration:       pomodoro.StudyDuration,
		DistractionDuration: pomodoro.DistractionDuration,
	}

	if err := h.PomodoroClockRepo.Create(context.Background(), pomodoroClock); err != nil {
		slog.Error("[MQTT] PomodoroHandler create pomodoro clock failed", "device_id", deviceID, "error", err)
		return
	}
}
