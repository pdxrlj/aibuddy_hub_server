package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"context"
	"errors"
)

// AgentRepo agetn仓库
type AgentRepo struct{}

// NewAgentRepo 创建agent仓库实例
func NewAgentRepo() *AgentRepo {
	return &AgentRepo{}
}

// GetAgentListByUID 获取agent
func (a *AgentRepo) GetAgentListByUID(ctx context.Context, uid int64, page, size int) ([]*model.Agent, int64, error) {
	_, span := tracer.Start(ctx, "GetAgentListByUID")
	defer span.End()
	offset := (page - 1) * size
	data, count, err := query.Agent.
		Where(query.Agent.UID.In(0, uid)).
		Order(query.Agent.CreatedAt.Desc()).FindByPage(offset, size)
	if err != nil {
		return nil, 0, nil
	}
	return data, count, err
}

// GetAgentByID 获取角色信息
func (a *AgentRepo) GetAgentByID(ctx context.Context, uid, roleID int64) error {
	_, span := tracer.Start(ctx, "GetAgentById")
	defer span.End()
	count, err := query.Agent.Where(query.Agent.ID.Eq(roleID), query.Agent.UID.In(uid, 0)).Count()

	if err != nil {
		return err
	}

	if count < 1 {
		return errors.New("角色信息异常")
	}

	return nil
}
