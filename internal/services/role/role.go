// Package role 服务层
package role

import (
	"aibuddy/internal/model"
	"aibuddy/internal/repository"
	"aibuddy/pkg/baidu"
	"aibuddy/pkg/config"
	"context"
	"errors"

	"github.com/cespare/xxhash/v2"
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
	RoleAPI    *baidu.Role
	SwitchRole *baidu.SwitchRole
}

// NewRoleService 实例化服务
func NewRoleService() *Service {
	return &Service{
		AgentRepo:  repository.NewAgentRepo(),
		DeviceRepo: repository.NewDeviceRepo(),
		RoleAPI:    baidu.NewRole(),
		SwitchRole: baidu.NewSwitchRole(),
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

// ChangeRoleName 切换设备角色
func (r *Service) ChangeRoleName(ctx context.Context, uid int64, deviceID string, roleName string) error {
	ctx, span := tracer().Start(ctx, "ChangeRoleName")
	defer span.End()

	if !r.DeviceRepo.CheckDeviceAuth(ctx, uid, deviceID) {
		return errors.New("无设置该设备的权限")
	}

	if err := r.DeviceRepo.ChangeDeviceRole(ctx, uid, deviceID, roleName); err != nil {
		return errors.New("切换角色失败")
	}
	instanceID := xxhash.Sum64String(deviceID)
	if err := r.SwitchRole.SwitchSceneRole(&baidu.SwitchRoleRequest{
		AiAgentInstanceID: instanceID, // 需要替换为有效的实例ID
		SceneRole:         roleName,
	}); err != nil {
		return errors.New("切换角色失败:" + err.Error())
	}

	return nil
}

// GetRoleByID 查看角色信息
func (r *Service) GetRoleByID(ctx context.Context, uid int64, roleID int64) (*model.Agent, error) {
	ctx, span := tracer().Start(ctx, "GetRoleByID")
	defer span.End()

	data, err := r.AgentRepo.GetAgentByID(ctx, uid, roleID)
	if err != nil {
		span.RecordError(errors.New("角色信息为空"))
		return nil, errors.New("角色信息为空")
	}

	return data, nil
}

// GetRoleListByAPI 通过API接口拉取角色列表
func (r *Service) GetRoleListByAPI(ctx context.Context) ([]*model.Agent, error) {
	_, span := tracer().Start(ctx, "GetRoleListByAPI")
	defer span.End()
	resp, err := r.RoleAPI.RoleList("")
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	result := make([]*model.Agent, 0, len(resp.LLM.Roles))
	for _, r := range resp.LLM.Roles {
		result = append(result, &model.Agent{
			ID:               r.ID,
			AgentName:        r.Name,
			RoleIntroduction: r.Description,
		})
	}

	return result, nil
}
