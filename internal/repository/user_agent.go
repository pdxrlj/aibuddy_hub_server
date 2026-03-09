// Package repository 数据库仓库层
package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"context"
	"errors"

	"gorm.io/gorm"
)

// UserAgentRepository 用户代理仓库
type UserAgentRepository struct{}

// NewUserAgentRepository 创建用户代理仓库
func NewUserAgentRepository() *UserAgentRepository {
	return &UserAgentRepository{}
}

// CreateUserAgent 创建 分析的用户使用代理的信息
func (r *UserAgentRepository) CreateUserAgent(ctx context.Context, userAgent *model.UserAgent) error {
	return query.UserAgent.WithContext(ctx).Create(userAgent)
}

// GetUserAgent 获取用户代理信息
func (r *UserAgentRepository) GetUserAgent(ctx context.Context, deviceID string, agentName string) (*model.UserAgent, error) {
	agent, err := query.UserAgent.WithContext(ctx).
		Where(
			query.UserAgent.DeviceID.Eq(deviceID),
			query.UserAgent.AgentName.Eq(agentName),
		).
		Order(query.UserAgent.CreatedAt.Desc()).
		First()

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("还没有你的数据呢,快去生成吧")
	}

	return agent, err
}
