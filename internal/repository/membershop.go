// Package repository 提供会员商城相关的数据库操作
package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gen"
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

// OrderList 获取订单列表
func (r *MemberShopRepository) OrderList(ctx context.Context, page, pageSize int, status ...string) ([]*model.Order, int64, error) {
	_, span := tracer.Start(ctx, "MemberShopRepository.OrderList")
	defer span.End()
	slog.Info("[Shop] OrderList", "page", page, "pageSize", pageSize, "status", status)
	offset := (page - 1) * pageSize
	orders, count, err := query.Order.
		Scopes(func(d gen.Dao) gen.Dao {
			if len(status) > 0 && status[0] != "" {
				return d.Where(query.Order.Status.In(status...))
			}
			return d
		}).
		Preload(query.Order.Goods.GoodsInfo).
		FindByPage(offset, pageSize)

	if err != nil {
		span.RecordError(err)
		span.SetAttributes(
			attribute.Int("page", page),
			attribute.Int("pageSize", pageSize),
		)
		return nil, 0, err
	}

	return orders, count, nil
}

// UpdateOrderStatus 更新订单状态（仅当订单为待支付状态时才更新）
func (r *MemberShopRepository) UpdateOrderStatus(ctx context.Context, outTradeNo string, transactionID string, status string) (int64, error) {
	_, span := tracer.Start(ctx, "MemberShopRepository.UpdateOrderStatus")
	defer span.End()

	result, err := query.Order.
		Where(query.Order.OutTradeNo.Eq(outTradeNo), query.Order.Status.Eq(string(model.OrderStatusPending))).
		Updates(map[string]any{
			query.Order.Status.ColumnName().String():        status,
			query.Order.TransactionID.ColumnName().String(): transactionID,
		})

	if err != nil {
		span.RecordError(err)
		span.SetAttributes(
			attribute.String("outTradeNo", outTradeNo),
			attribute.String("transactionId", transactionID),
			attribute.String("status", status),
		)
		return 0, err
	}

	return result.RowsAffected, nil
}

// UpdateOrderStatusToRefunded 更新订单状态为已退款
func (r *MemberShopRepository) UpdateOrderStatusToRefunded(ctx context.Context, outTradeNo string) error {
	_, span := tracer.Start(ctx, "MemberShopRepository.UpdateOrderStatusToRefunded")
	defer span.End()

	_, err := query.Order.
		Where(query.Order.OutTradeNo.Eq(outTradeNo)).
		Update(query.Order.Status, model.OrderStatusRefunded)

	if err != nil {
		span.RecordError(err)
		span.SetAttributes(
			attribute.String("outTradeNo", outTradeNo),
		)
		return err
	}

	return nil
}

// GetOrderByOutTradeNo 根据商户订单号查询订单详情
func (r *MemberShopRepository) GetOrderByOutTradeNo(ctx context.Context, outTradeNo string) (*model.Order, error) {
	_, span := tracer.Start(ctx, "MemberShopRepository.GetOrderByOutTradeNo")
	defer span.End()

	order, err := query.Order.
		Where(query.Order.OutTradeNo.Eq(outTradeNo)).
		Preload(query.Order.Goods).
		Preload(query.Order.Goods.GoodsInfo).
		First()

	if err != nil {
		span.RecordError(err)
		span.SetAttributes(
			attribute.String("outTradeNo", outTradeNo),
		)
		return nil, err
	}

	return order, nil
}
