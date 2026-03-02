// Package role 服务层
package role

import (
	"aibuddy/internal/model"
	"aibuddy/internal/repository"
	"aibuddy/pkg/config"
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var tracer = func() trace.Tracer {
	return otel.Tracer(config.Instance.Tracer.ServiceName)
}

// Service 角色服务
type Service struct {
	AgentRepo  *repository.AgentRepo
	DeviceRepo *repository.DeviceRepo
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

// ChangeRole 切换设备角色
func (r *Service) ChangeRole(ctx context.Context, uid int64, deviceID string, roleID int64) error {
	ctx, span := tracer().Start(ctx, "ChangeRoleById")
	defer span.End()

	if !r.AgentRepo.ChcekAgentByID(ctx, uid, roleID) {
		return errors.New("role_id参数异常")
	}

	_ = deviceID
	// if err := r.DeviceRepo.ChangeDeviceRole(ctx, uid, deviceID, roleID); err != nil {
	// 	return errors.New("device_id参数异常")
	// }

	return nil
}

// GetRoleByID 查看角色信息
func (r *Service) GetRoleByID(ctx context.Context, uid int64, roleID int64) (*model.Agent, error) {
	ctx, span := tracer().Start(ctx, "GetRoleByID")
	defer span.End()

	data, err := r.AgentRepo.GetAgentByID(ctx, uid, roleID)
	if err != nil {
		return nil, errors.New("角色信息为空")
	}

	return data, nil
}
