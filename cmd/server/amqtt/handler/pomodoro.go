// Package handler is the pomodoro service
package handler

import (
	"aibuddy/aiframe/pomodoro"
	"aibuddy/internal/model"
	"aibuddy/internal/repository"
	"aibuddy/pkg/mqtt"
	"context"
	"encoding/json"
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
		DeviceID:      deviceID,
		TotalDuration: pomodoro.TotalTime,
		StudyDuration: pomodoro.StudyDuration,
	}

	// 序列化分心记录
	if pomodoro.DistractionRecord != nil {
		record := model.DistractionRecord{
			TouchScreenCount: pomodoro.DistractionRecord.TouchScreenCount,
			TouchHead:        pomodoro.DistractionRecord.TouchHead,
		}
		recordJSON, err := json.Marshal(record)
		if err != nil {
			slog.Error("[MQTT] PomodoroHandler marshal distraction record failed", "device_id", deviceID, "error", err)
		} else {
			pomodoroClock.DistractionRecord = recordJSON
		}
	}

	if err := h.PomodoroClockRepo.Create(context.Background(), pomodoroClock); err != nil {
		slog.Error("[MQTT] PomodoroHandler create pomodoro clock failed", "device_id", deviceID, "error", err)
		return
	}
}
