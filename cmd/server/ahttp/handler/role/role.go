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
func (r *Handler) RoleList(state *ahttp.State, req *ListRequest) error {
	_, span := tracer().Start(state.Ctx.Request().Context(), "role_list")
	defer span.End()

	span.SetAttributes(attribute.Int("page", req.Page))
	span.SetAttributes(attribute.Int("size", req.Size))

	uid, err := auth.GetUIDFromContext(state.Ctx)
	if err != nil {
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	data, count, err := r.RoleSerivce.GetRoleListByUID(state.Context(), uid, req.Page, req.Size)
	if err != nil {
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	var res []RolesResponse
	for _, v := range data {
		t := &RolesResponse{
			ID:               v.ID,
			UID:              v.UID,
			AgentName:        v.AgentName,
			RoleIntroduction: v.RoleIntroduction,
			SystemPrompt:     v.SystemPrompt,
			CreatedAt:        v.CreatedAt,
			UpdatedAt:        v.UpdatedAt,
		}

		res = append(res, *t)
	}

	return state.Resposne().SetData(ListResponse{
		Total: int(count),
		Page:  req.Page,
		Size:  req.Size,
		Roles: res,
	}).Success()
}
