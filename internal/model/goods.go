// Package model 商品模型
package model

// Goods 商品
type Goods struct {
	ID    int64  `gorm:"column:id;autoIncrement;primaryKey" json:"id"`
	Name  string `gorm:"column:name;type:varchar(255);comment:商品名称" json:"name"`
	Price int64  `gorm:"column:price;type:int;comment:商品价格" json:"price"`
	Stock int64  `gorm:"column:stock;type:int;comment:商品库存" json:"stock"`
	// 活动价格
	ActivityPrice int64 `gorm:"column:activity_price;type:int;comment:活动价格" json:"activity_price"`

	// 商品描述
	Description string `gorm:"column:description;type:text;comment:商品描述" json:"description"`

	CreatedAt LocalTime `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt LocalTime `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
}

// TableName 表名
func (g *Goods) TableName() string {
	return TableName("goods")
}
