// Package role 服务层
package role

import (
	"aibuddy/internal/model"
	"aibuddy/internal/repository"
	"aibuddy/pkg/config"
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var tracer = func() trace.Tracer {
	return otel.Tracer(config.Instance.Tracer.ServiceName)
}

// Service 角色服务
type Service struct {
	AgentRepo *repository.AgentRepo
}

// NewRoleService 实例化服务
func NewRoleService() *Service {
	return &Service{
		AgentRepo: repository.NewAgentRepo(),
	}
}

// GetRoleListByUID 获取role列表
func (r *Service) GetRoleListByUID(ctx context.Context, uid int64, page int, size int) ([]*model.Agent, int64, error) {
	ctx, span := tracer().Start(ctx, "UpsertUser")
	defer span.End()
	data, count, err := r.AgentRepo.GetAgentListByUID(ctx, uid, page, size)

	if err != nil {
		return nil, 0, err
	}
	return data, count, nil
}
