// Package role handler层
package role

import (
	aiuserService "aibuddy/internal/services/aiuser"
	"aibuddy/internal/services/role"
	"aibuddy/pkg/ahttp"
	"aibuddy/pkg/config"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = func() trace.Tracer {
	return otel.Tracer(config.Instance.Tracer.ServiceName)
}

// Handler handler
type Handler struct {
	RoleSerivce *role.Service
}

// NewRoleHandler 实例化handler
func NewRoleHandler() *Handler {
	return &Handler{
		RoleSerivce: role.NewRoleService(),
	}
}

// RoleList 角色列表
func (r *Handler) RoleList(state *ahttp.State) error {
	ctx, span := tracer().Start(state.Ctx.Request().Context(), "role_list")
	defer span.End()

	span.SetAttributes(attribute.Int("page", req.Page))
	span.SetAttributes(attribute.Int("size", req.Size))

	uid, err := aiuserService.GetUIDFromContext(state.Ctx)
	if err != nil {
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	data, count, err := r.RoleSerivce.GetRoleListByUID(state.Context(), uid, req.Page, req.Size)
	if err != nil {
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}
	var res []RolesResponse
	for _, v := range data {
		res = append(res, RolesResponse{
			ID:               v.ID,
			AgentName:        v.AgentName,
			RoleIntroduction: v.RoleIntroduction,
		})
	}

	return state.Resposne().SetData(ListResponse{
		Total: len(res),
		Page:  1,
		Size:  10,
		Roles: res,
	}).Success()
}

// ChangeRole 切换角色信息
func (r *Handler) ChangeRole(state *ahttp.State, req *ChangeRquest) error {
	ctx, span := tracer().Start(state.Ctx.Request().Context(), "change_role")
	defer span.End()
	span.SetAttributes(attribute.String("device_id", req.DeviceID))
	span.SetAttributes(attribute.String("role_name", req.RoleName))

	uid, err := aiuserService.GetUIDFromContext(state.Ctx)
	if err != nil {
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	err = r.RoleSerivce.ChangeRoleName(ctx, uid, req.DeviceID, req.RoleName)
	if err != nil {
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	return state.Resposne().Success()
}

// RoleInfo 获取角色信息
func (r *Handler) RoleInfo(state *ahttp.State, req *InfoRequest) error {
	ctx, span := tracer().Start(state.Ctx.Request().Context(), "change_role")
	defer span.End()
	span.SetAttributes(attribute.Int("role_id", int(req.RoleID)))

	uid, err := aiuserService.GetUIDFromContext(state.Ctx)
	if err != nil {
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}
	data, err := r.RoleSerivce.GetRoleByID(ctx, uid, req.RoleID)
	if err != nil {
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	return state.Resposne().SetData(&InfoResponse{
		ID:               data.ID,
		AgentName:        data.AgentName,
		UID:              data.UID,
		RoleIntroduction: data.RoleIntroduction,
		SystemPrompt:     data.SystemPrompt,
		CreatedAt:        data.CreatedAt.Format(time.DateTime),
		UpdatedAt:        data.UpdatedAt.Format(time.DateTime),
	}).Success()
}
