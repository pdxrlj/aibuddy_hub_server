// Package repository 提供数据库操作封装
package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"context"
)

// GrowthReportRepo 成长报告仓库
type GrowthReportRepo struct {
}

// NewGrowthReportRepo 创建成长报告仓库实例
func NewGrowthReportRepo() *GrowthReportRepo {
	return &GrowthReportRepo{}
}

// Create 创建成长报告
func (g *GrowthReportRepo) Create(ctx context.Context, report *model.GrowthReport) error {
	_, span := tracer.Start(ctx, "GrowthReportRepo.Create")
	defer span.End()

	return query.GrowthReport.Create(report)
}

// GetListByDeviceID 分页获取指定设备的成长报告
func (g *GrowthReportRepo) GetListByDeviceID(ctx context.Context, deviceID string, page, size int) ([]*model.GrowthReport, int64, error) {
	_, span := tracer.Start(ctx, "GrowthReportRepo.GetListByDeviceID")
	defer span.End()

	offset := (page - 1) * size
	return query.GrowthReport.WithContext(ctx).
		Where(query.GrowthReport.DeviceID.Eq(deviceID)).
		Order(query.GrowthReport.StartTime.Desc()).
		FindByPage(offset, size)
}
