// Package membershop 会员商城
package membershop

// GoodsListRequest 获取商品列表请求
type GoodsListRequest struct {
	Page     int `json:"page" query:"page" validate:"required,min=1"`
	PageSize int `json:"page_size" query:"page_size" validate:"required,min=1"`
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	GoodsID int64 `json:"goods_id" query:"goods_id" validate:"required,min=1"`
}
