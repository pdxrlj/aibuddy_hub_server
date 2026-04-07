// Package model 订单商品模型
package model

// OrderGoods 订单商品
type OrderGoods struct {
	ID      int64 `gorm:"column:id;autoIncrement;primaryKey" json:"id"`
	OrderID int64 `gorm:"column:order_id;type:int;comment:订单ID;index" json:"order_id"`
	GoodsID int64 `gorm:"column:goods_id;type:int;comment:商品ID;index" json:"goods_id"`

	GoodsInfo Goods `gorm:"foreignKey:GoodsID;references:ID" json:"goods_info"`

	GoodsName  string    `gorm:"column:goods_name;type:varchar(255);comment:商品名称;index" json:"goods_name"`
	GoodsPrice int64     `gorm:"column:goods_price;type:int;comment:商品价格;index" json:"goods_price"`
	GoodsNum   int64     `gorm:"column:goods_num;type:int;comment:商品数量;index" json:"goods_num"`
	CreatedAt  LocalTime `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt  LocalTime `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
}

// TableName 表名
func (o *OrderGoods) TableName() string {
	return TableName("order_goods")
}
