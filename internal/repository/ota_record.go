package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"aibuddy/pkg/helpers"
	"context"

	"gorm.io/gorm"
)

// OtaResourceRepo OTA资源仓库
type OtaResourceRepo struct{}

// NewOtaResourceRepo 创建OTA资源仓库
func NewOtaResourceRepo() *OtaResourceRepo {
	return &OtaResourceRepo{}
}

// GetLatestOtaResource 获取最新的OTA资源
func (r *OtaResourceRepo) GetLatestOtaResource(ctx context.Context, boardType, currentVersion string) (*model.OtaResource, error) {
	resources, err := query.OtaResource.WithContext(ctx).
		Order(query.OtaResource.ID.Desc()).
		Where(query.OtaResource.BoardType.Eq(boardType)).
		Find()
	if err != nil {
		return nil, err
	}

	for _, res := range resources {
		if helpers.CompareVersion(res.Version, currentVersion) > 0 {
			return res, nil
		}
	}

	return nil, gorm.ErrRecordNotFound
}
