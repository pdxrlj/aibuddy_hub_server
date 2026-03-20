package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"context"
)

// OtaResourceRepo OTA资源仓库
type OtaResourceRepo struct{}

// NewOtaResourceRepo 创建OTA资源仓库
func NewOtaResourceRepo() *OtaResourceRepo {
	return &OtaResourceRepo{}
}

// GetLatestOtaResource 获取最新的OTA资源
func (r *OtaResourceRepo) GetLatestOtaResource(ctx context.Context) (*model.OtaResource, error) {
	otaResource, err := query.OtaResource.WithContext(ctx).
		Order(query.OtaResource.ID.Desc()).
		First()
	if err != nil {
		return nil, err
	}
	return otaResource, nil
}
