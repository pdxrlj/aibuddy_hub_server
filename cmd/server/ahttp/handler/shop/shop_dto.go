// Package shop 商城请求体
package shop

import "aibuddy/pkg/shop"

// GoodsListRequest 商品列表请求
type GoodsListRequest struct {
	PageSize int    `json:"page_size" form:"page_size" param:"page_size" query:"page_size" validate:"gte=1" default:"10"`
	NextKey  string `json:"next_key" validate:"omitempty"`
}

// GoodsListResponse 商品列表相应
type GoodsListResponse struct {
	Total   int                            `json:"total"`
	NextKey string                         `json:"next_key"`
	List    []*shop.ProductProductResponse `json:"list"`
}
