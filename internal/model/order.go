// Package model 订单模型
package model

import (
	"time"

	"gorm.io/gorm"
)

// DefaultExpireTime 默认过期时间 10分钟
const DefaultExpireTime = time.Minute * 10

// OrderStatus 订单状态
type OrderStatus string

const (
	// OrderStatusPending 待支付
	OrderStatusPending OrderStatus = "待支付"
	// OrderStatusPaid 已支付
	OrderStatusPaid OrderStatus = "已支付"
	// OrderStatusTimeout 已超时
	OrderStatusTimeout OrderStatus = "已超时"
	// OrderStatusRefunded 已退款
	OrderStatusRefunded OrderStatus = "已退款"
)

// String 返回订单状态字符串
func (o OrderStatus) String() string {
	return string(o)
}

// Order 订单
type Order struct {
	ID         int64  `gorm:"column:id;autoIncrement;primaryKey" json:"id"`
	UserID     int64  `gorm:"column:user_id;index;comment:用户ID" json:"user_id"`
	DeviceID   string `gorm:"column:device_id;index;comment:设备ID" json:"device_id"`
	ActivityID int64  `gorm:"column:activity_id;index;comment:活动ID" json:"activity_id"`

	OutTradeNo    string `gorm:"column:out_trade_no;uniqueIndex;type:varchar(64);comment:商户订单号" json:"out_trade_no"`
	TransactionID string `gorm:"column:transaction_id;index;type:varchar(64);comment:微信支付订单号" json:"transaction_id"`

	Status OrderStatus `gorm:"column:status;index;comment:订单状态" json:"status"`

	Goods        []*OrderGoods  `gorm:"foreignKey:OrderID;references:ID" json:"goods"`
	ActivityInfo *GoodsActivity `gorm:"foreignKey:ActivityID;references:ID" json:"activity_info"`

	ExpireTime LocalTime `gorm:"column:expire_time;comment:订单超时时间" json:"expire_time"`

	CreatedAt LocalTime `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt LocalTime `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
}

// TableName 表名
func (o *Order) TableName() string {
	return TableName("order")
}

// IsPayTimeExpired 订单支付时间是否已经过期
func (o *Order) IsPayTimeExpired() bool {
	return o.Status == OrderStatusPending && !o.ExpireTime.IsZero() && o.ExpireTime.Before(NowLocal())
}

// BeforeCreate 创建前钩子
func (o *Order) BeforeCreate(_ *gorm.DB) error {
	now := NowLocal()
	o.ExpireTime = LocalTime(now.Time().Add(DefaultExpireTime))
	o.Status = OrderStatusPending
	o.CreatedAt = now
	o.UpdatedAt = now
	return nil
}

// BeforeUpdate 更新前钩子
func (o *Order) BeforeUpdate(_ *gorm.DB) error {
	o.UpdatedAt = NowLocal()
	return nil
}

// AfterFind 查询之后，判断订单是否过期
func (o *Order) AfterFind(tx *gorm.DB) error {
	if o.Status == OrderStatusPending && !o.ExpireTime.IsZero() {
		if o.ExpireTime.Before(NowLocal()) {
			o.Status = OrderStatusTimeout
			tx.Model(o).Update("status", OrderStatusTimeout)
		}
	}

	return nil
}
