package model

import "time"

// GoodsActivity 商品活动
type GoodsActivity struct {
	ID        int64     `gorm:"column:id;autoIncrement;primaryKey" json:"id"`
	GoodsID   int64     `gorm:"colum:goods_id;type:int;comment:商品ID" json:"gid"`
	Pirce     int64     `gorm:"column:price;type:int;comment:活动价格" json:"price"`
	GoodsInfo *Goods    `gorm:"foreignKey:GoodsID;references:ID" json:"goods_info,omitempty"`
	Type      int       `gorm:"column:type;type:int;comment:活动类型" json:"type"`
	Count     int64     `gorm:"column:count;type:int;comment:活动数量" json:"count"`
	Title     string    `gorm:"column:title;type:varchar(255);comment:活动标题" json:"title"`
	Desc      string    `gorm:"column:desc;type:text;comment:活动描述" json:"desc"`
	StartTime time.Time `gorm:"column:start_time;comment:活动开始时间" json:"start_time"`
	EndTime   time.Time `gorm:"column:end_time;comment:活动结束时间" json:"end_time"`
}

// TableName 表名
func (g *GoodsActivity) TableName() string {
	return TableName("goods_activity")
}
