// Package repository 提供会员商城相关的数据库操作
package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"context"

	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
)

// MemberShopRepository 会员商城仓库
type MemberShopRepository struct {
}

// NewMemberShopRepository 实例化
func NewMemberShopRepository() *MemberShopRepository {
	return &MemberShopRepository{}
}

// GoodsList 获取商品列表
func (r *MemberShopRepository) GoodsList(ctx context.Context, page, pageSize int) ([]*model.Goods, int64, error) {
	_, span := tracer.Start(ctx, "MemberShopRepository.GoodsList")

	offset := (page - 1) * pageSize
	goods, count, err := query.Goods.FindByPage(offset, pageSize)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(
			attribute.Int("page", page),
			attribute.Int("pageSize", pageSize),
		)
		return nil, 0, err
	}

	return goods, count, err
}

// GoodsInfo 查询商品信息
func (r *MemberShopRepository) GoodsInfo(ctx context.Context, goodsID int64) (*model.Goods, error) {
	_, span := tracer.Start(ctx, "MemberShopRepository.GoodsInfo")
	defer span.End()

	goods, err := query.Goods.Where(query.Goods.ID.Eq(goodsID)).First()
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(
			attribute.Int64("goodsID", goodsID),
		)
		return nil, err
	}

	return goods, nil
}

// Transaction 执行事务
func (r *MemberShopRepository) Transaction(fn func(tx *gorm.DB) error) error {
	return query.Q.Transaction(func(tx *query.Query) error {
		return fn(tx.UnderlyingDB())
	})
}
