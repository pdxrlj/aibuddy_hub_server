// Package role handler层
package role

import (
	"aibuddy/internal/services/auth"
	"aibuddy/internal/services/role"
	"aibuddy/pkg/ahttp"
	"aibuddy/pkg/config"
	"net/http"

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

	data, err := r.RoleSerivce.GetRoleListByAPI(ctx)
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

	uid, err := auth.GetUIDFromContext(state.Ctx)
	if err != nil {
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	err = r.RoleSerivce.ChangeRoleName(ctx, uid, req.DeviceID, req.RoleName)
	if err != nil {
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	return state.Resposne().Success()
}
