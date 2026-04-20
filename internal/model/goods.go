// Package model 商品模型
package model

// GoodsStatus 商品状态
type GoodsStatus string

const (
	// GoodsStatusNormal 正常状态
	GoodsStatusNormal GoodsStatus = "正常"
	// GoodsStatusSold 售罄状态
	GoodsStatusSold GoodsStatus = "售罄"
)

// String 转换为字符串
func (g GoodsStatus) String() string {
	return string(g)
}

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

	// 商品状态
	Status     GoodsStatus `gorm:"column:status;type:string;comment:商品状态" json:"status"`
	UsageLimit int64       `gorm:"column:usage_limit;type:int;default:0;comment:可使用次数" json:"usage_limit"`

	CreatedAt LocalTime `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt LocalTime `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
}

// TableName 表名
func (g *Goods) TableName() string {
	return TableName("goods")
}
