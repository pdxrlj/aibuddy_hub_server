// Package repository 订单相关数据库操作
package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

// OrderRepo 订单仓库
type OrderRepo struct{}

// NewOrderRepo 创建订单仓库实例
func NewOrderRepo() *OrderRepo {
	return &OrderRepo{}
}

// GetMemberLevel 会员等级0未开通 1已开通 2已过期
func (o *OrderRepo) GetMemberLevel(ctx context.Context, userID int64, deviceID string) int {
	_, span := tracer.Start(ctx, "ShopService.GetMemberLevel")
	defer span.End()
	member, err := query.Order.Where(query.Order.UserID.Eq(userID), query.Order.DeviceID.Eq(deviceID), query.Order.Status.Eq(model.OrderStatusPaid.String())).First()
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.Int64("userID", userID), attribute.String("deviceID", deviceID))
		return 0
	}

	if member.ExpireTime.Unix() > time.Now().Unix() {
		return 1
	}
	return 2
}
