// Package repository 提供会员商城相关的数据库操作
package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"context"
	"errors"
	"log/slog"
	"time"

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
func (r *MemberShopRepository) GoodsList(ctx context.Context, page, pageSize int) ([]*model.GoodsActivity, int64, error) {
	_, span := tracer.Start(ctx, "MemberShopRepository.GoodsList")

	offset := (page - 1) * pageSize
	goods, count, err := query.GoodsActivity.Preload(query.GoodsActivity.GoodsInfo).FindByPage(offset, pageSize)
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

// GetActvityByGoodsID 根据商品ID查询活动信息
func (r *MemberShopRepository) GetActvityByGoodsID(ctx context.Context, goodsID int64) ([]*model.GoodsActivity, error) {
	_, span := tracer.Start(ctx, "MemberShopRepository.GetActvityByGoodsID")
	defer span.End()

	actvityList, err := query.GoodsActivity.
		Where(query.GoodsActivity.GoodsID.Eq(goodsID),
			query.GoodsActivity.StartTime.Lte(time.Now()),
			query.GoodsActivity.EndTime.Gte(time.Now()),
		).Preload(query.GoodsActivity.GoodsInfo).Find()
	if err != nil {
		return nil, err
	}
	return actvityList, nil
}

// SubActivityCount 活动购买成功后，扣减活动库存
func (r *MemberShopRepository) SubActivityCount(ctx context.Context, aid int64) error {
	_, span := tracer.Start(ctx, "MemberShopRepository.SubActivityCount")
	defer span.End()
	_, err := query.GoodsActivity.Where(query.GoodsActivity.ID.Eq(aid), query.GoodsActivity.Count_.Gte(0)).Update(query.GoodsActivity.Count_, gorm.Expr("count - 1"))
	return err
}

// GetGoodsByName 根据名称查询商品列表
func (r *MemberShopRepository) GetGoodsByName(ctx context.Context, name string) ([]*model.Goods, error) {
	_, span := tracer.Start(ctx, "MemberShopRepository.SubActivityCount")
	defer span.End()
	produets, err := query.Goods.Where(query.Goods.Name.Eq(name)).Find()
	if err != nil {
		return nil, err
	}
	return produets, nil
}

// GetMemberByDeviceID 根据设备ID查询会员信息
func (r *MemberShopRepository) GetMemberByDeviceID(ctx context.Context, deviceID string) (*model.Order, error) {
	_, span := tracer.Start(ctx, "MemberShopRepository.GetMemberByDeviceID")
	defer span.End()

	member, err := query.Order.Where(query.Order.DeviceID.Eq(deviceID)).
		Where(query.Order.Status.Eq(model.OrderStatusPaid.String())).
		Preload(query.Order.Goods).
		Preload(query.Order.Goods.GoodsInfo).
		Order(query.Order.ExpireTime.Desc()).
		First()

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return member, nil
}

// GetMemberRoleNum 根据设备ID查询会员拥有的角色数量
func (r *MemberShopRepository) GetMemberRoleNum(ctx context.Context, deviceID string) int {
	_, span := tracer.Start(ctx, "MemberShopRepository.GetMemberRoleNum")
	defer span.End()
	num := 0

	result, err := query.Device.Where(query.Device.DeviceID.Eq(deviceID)).First()
	if err != nil {
		return num
	}
	if result != nil {
		num = result.SurplusNum
	}

	return num
}
