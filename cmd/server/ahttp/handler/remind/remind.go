// Package remindhandler  提醒
package remindhandler

import (
	"aibuddy/internal/model"
	aiuserService "aibuddy/internal/services/aiuser"
	"aibuddy/internal/services/remind"
	"aibuddy/pkg/ahttp"
	"aibuddy/pkg/config"
	"errors"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = func() trace.Tracer {
	return otel.Tracer(config.Instance.Tracer.ServiceName)
}

// Remind remind handler
type Remind struct {
	RemindService *remind.Service
}

// NewRemindHandler new remind handler
func NewRemindHandler() *Remind {
	return &Remind{
		RemindService: remind.NewRemindService(),
	}
}

// CreateRemind 添加提醒事件
func (m *Remind) CreateRemind(state *ahttp.State, req *AddRemindRequest) error {
	ctx, span := tracer().Start(state.Context(), "CreateRemind")
	defer span.End()

	uid, err := aiuserService.GetUIDFromContext(state.Ctx)
	if err != nil {
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	reminderTime, err := time.Parse(time.DateTime, req.ReminderTime)
	if err != nil {
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(errors.New("时间格式错误"))
	}

	if err := m.RemindService.SubmitRemind(ctx, uid, &model.Reminder{
		RepeatType:      model.RepeatType(req.RepeatType),
		ReminderTitle:   req.ReminderTitle,
		ReminderContent: req.ReminderContent,
		ReminderTime:    reminderTime,
		DeviceID:        req.DeviceID,
		Status:          model.ReminderStatus(req.Status),
	}); err != nil {
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	return state.Resposne().Success()
}

// UpdateRemind 更新数据
func (m *Remind) UpdateRemind(state *ahttp.State, req *AddRemindRequest) error {
	ctx, span := tracer().Start(state.Context(), "UpdateRemind")
	defer span.End()
	uid, err := aiuserService.GetUIDFromContext(state.Ctx)
	if err != nil {
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}
	reminderTime, err := time.Parse(time.DateTime, req.ReminderTime)
	if err != nil {
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(errors.New("时间格式错误"))
	}
	if req.ID <= 0 {
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(errors.New("缺少必要参数ID"))
	}

	if err := m.RemindService.SubmitRemind(ctx, uid, &model.Reminder{
		ID:              req.ID,
		RepeatType:      model.RepeatType(req.RepeatType),
		ReminderTitle:   req.ReminderTitle,
		ReminderContent: req.ReminderContent,
		ReminderTime:    reminderTime,
		DeviceID:        req.DeviceID,
		Status:          model.ReminderStatus(req.Status),
	}); err != nil {
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}
	return state.Resposne().Success()
}

// DeleteRemind 删除提醒事件
func (m *Remind) DeleteRemind(state *ahttp.State, req *RemindRequest) error {
	ctx, span := tracer().Start(state.Context(), "DeleteRemind")
	defer span.End()

	uid, err := aiuserService.GetUIDFromContext(state.Ctx)
	if err != nil {
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}
	if err := m.RemindService.DeleateRemindByID(ctx, uid, req.ID); err != nil {
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	return state.Resposne().Success()
}

// ListRemind 提醒事件列表
func (m *Remind) ListRemind(state *ahttp.State, req *ListReqeust) error {
	ctx, span := tracer().Start(state.Context(), "DeleteRemind")
	defer span.End()

	span.SetAttributes(attribute.Int("page", int(req.Page)))
	span.SetAttributes(attribute.Int("size", int(req.Size)))

	uid, err := aiuserService.GetUIDFromContext(state.Ctx)
	if err != nil {
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	data, total, err := m.RemindService.GetList(ctx, uid, req.DeviceID, req.Page, req.Size)
	if err != nil {
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	result := &ListResponse{Total: total, Page: req.Page, Size: req.Size, Data: make([]*ReminderInfo, 0)}

	for _, v := range data {
		result.Data = append(result.Data, &ReminderInfo{
			ID:              v.ID,
			DeviceID:        req.DeviceID,
			RepeatType:      v.RepeatType.String(),
			ReminderTitle:   v.ReminderTitle,
			ReminderContent: v.ReminderContent,
			ReminderTime:    time.Unix(v.ReminderTime.Unix(), 0).Format(time.DateTime),
			Status:          v.Status.String(),
		})
	}

	return state.Resposne().SetData(result).Success()
}
