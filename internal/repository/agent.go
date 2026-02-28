package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"context"
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
		Debug().
		Where(query.Agent.UID.In(0, uid)).
		Order(query.Agent.CreatedAt.Desc()).FindByPage(offset, size)
	if err != nil {
		return nil, 0, nil
	}
	return data, count, err
}
