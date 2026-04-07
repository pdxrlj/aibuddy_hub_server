// Package anniversaryhandler 纪念日handler层
package anniversaryhandler

import (
	"aibuddy/internal/model"
	aiuserService "aibuddy/internal/services/aiuser"
	"aibuddy/internal/services/anniversary"
	"aibuddy/pkg/ahttp"
	"aibuddy/pkg/config"
	"errors"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var tracer = func() trace.Tracer {
	return otel.Tracer(config.Instance.Tracer.ServiceName)
}

// Anniversary Anniversary结构
type Anniversary struct {
	AnniversaryServer *anniversary.Service
}

// NewAnniversaryHandler 实例化NewAnniversaryHandler
func NewAnniversaryHandler() *Anniversary {
	return &Anniversary{
		AnniversaryServer: anniversary.NewAnniversaryService(),
	}
}

// CreateAnniversary 创建纪念日列表
func (a *Anniversary) CreateAnniversary(state *ahttp.State, r *AnniversaryInfoCreateRequest) error {
	ctx, span := tracer().Start(state.Context(), "CreateAnniversary")
	defer span.End()

	uid, err := aiuserService.GetUIDFromContext(state.Ctx)
	if err != nil {
		return state.Response().SetStatus(http.StatusBadRequest).Error(err)
	}

	anniversaryTime, err := time.Parse(time.DateTime, r.AnniversaryTime)
	if err != nil {
		return state.Response().Error(errors.New("纪念日时间格式错误"))
	}

	if err := a.AnniversaryServer.SubmitAnniversary(ctx, uid, r.DeviceID, &model.AnniversaryReminder{
		DeviceID:         r.DeviceID,
		AnniversaryType:  model.AnniversaryType(r.AnniversaryType),
		ReminderUsername: r.ReminderUsername,
		ReminderUserSex:  r.ReminderUserSex,
		AnniversaryTime:  anniversaryTime,
		ReminderWay:      model.ReminderWay(r.ReminderWay),
		Remarks:          r.Remarks,
	}); err != nil {
		return state.Response().Error(errors.New("创建纪念日失败:" + err.Error()))
	}

	return state.Response().Success()
}

// UpdateAnniversary 创建纪念日列表
func (a *Anniversary) UpdateAnniversary(state *ahttp.State, r *AnniversaryInfoUpdateRequest) error {
	ctx, span := tracer().Start(state.Context(), "UpdateAnniversary")
	defer span.End()

	uid, err := aiuserService.GetUIDFromContext(state.Ctx)
	if err != nil {
		return state.Response().SetStatus(http.StatusBadRequest).Error(err)
	}

	anniversaryTime, err := time.Parse(time.DateTime, r.AnniversaryTime)
	if err != nil {
		return state.Response().Error(errors.New("纪念日时间格式错误"))
	}

	if err := a.AnniversaryServer.SubmitAnniversary(ctx, uid, r.DeviceID, &model.AnniversaryReminder{
		ID:               r.ID,
		DeviceID:         r.DeviceID,
		AnniversaryType:  model.AnniversaryType(r.AnniversaryType),
		ReminderUsername: r.ReminderUsername,
		ReminderUserSex:  r.ReminderUserSex,
		AnniversaryTime:  anniversaryTime,
		ReminderWay:      model.ReminderWay(r.ReminderWay),
		Remarks:          r.Remarks,
	}); err != nil {
		return state.Response().Error(errors.New("更新纪念日失败:" + err.Error()))
	}

	return state.Response().Success()
}

// DeleateAnniversary 删除纪念日列表
func (a *Anniversary) DeleateAnniversary(state *ahttp.State, r *InfoRequest) error {
	ctx, span := tracer().Start(state.Context(), "CreateAnniversary")
	defer span.End()

	uid, err := aiuserService.GetUIDFromContext(state.Ctx)
	if err != nil {
		return state.Response().SetStatus(http.StatusBadRequest).Error(err)
	}

	if err := a.AnniversaryServer.DeleteAnniversary(ctx, uid, r.DeviceID, r.ID); err != nil {
		return state.Response().Error(errors.New("删除纪念日失败:" + err.Error()))
	}

	return state.Response().Success()
}

// ListAnniversary 删除纪念日列表
func (a *Anniversary) ListAnniversary(state *ahttp.State, r *ListRequest) error {
	ctx, span := tracer().Start(state.Context(), "CreateAnniversary")
	defer span.End()

	uid, err := aiuserService.GetUIDFromContext(state.Ctx)
	if err != nil {
		return state.Response().SetStatus(http.StatusBadRequest).Error(err)
	}

	data, total, err := a.AnniversaryServer.GetListByPage(ctx, uid, r.DeviceID, r.Page, r.Size)
	if err != nil {
		return state.Response().SetStatus(http.StatusBadRequest).Error(err)
	}

	result := make([]*InfoReponse, 0, len(data))
	for _, v := range data {
		result = append(result, &InfoReponse{
			ID:               v.ID,
			DeviceID:         v.DeviceID,
			AnniversaryType:  string(v.AnniversaryType),
			ReminderUsername: v.ReminderUsername,
			ReminderUserSex:  v.ReminderUserSex,
			AnniversaryTime:  v.AnniversaryTime.Format(time.DateOnly),
			ReminderWay:      v.ReminderWay.String(),
			Remarks:          v.Remarks,
		})
	}

	return state.Response().SetData(ListReponse{
		Total: total,
		Page:  r.Page,
		Size:  r.Size,
		Data:  result,
	}).Success()
}
