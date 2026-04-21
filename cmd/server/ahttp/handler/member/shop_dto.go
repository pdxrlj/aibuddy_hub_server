// Package membershop 会员商城
package membershop

// GoodsListRequest 获取商品列表请求
type GoodsListRequest struct {
	Page     int `json:"page" query:"page" validate:"required,min=1"`
	PageSize int `json:"page_size" query:"page_size" validate:"required,min=1"`
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	GoodsID  int64  `json:"goods_id" query:"goods_id" validate:"required,min=1"`
	DeviceID string `json:"device_id" query:"device_id" validate:"required"`
}

// OrderPayRequest 待支付订单支付请求
type OrderPayRequest struct {
	OrderID string `json:"order_id" query:"order_id" validate:"required"`
}

// PaySuccessRequest 支付成功请求
type PaySuccessRequest struct{}

// SkipBodyBind 跳过参数绑定
func (p *PaySuccessRequest) SkipBodyBind() {}

// RefundSuccessRequest 退款成功请求
type RefundSuccessRequest struct{}

// SkipBodyBind 跳过参数绑定
func (r *RefundSuccessRequest) SkipBodyBind() {}

// OrderListRequest 获取订单列表请求
type OrderListRequest struct {
	Page     int    `json:"page" query:"page" validate:"required,min=1"`
	PageSize int    `json:"page_size" query:"page_size" validate:"required,min=1"`
	Status   string `json:"status" query:"status" validate:"omitempty,oneof=待支付 已支付 已超时 已退款 全部" msg:"oneof:订单状态无效"`
}

// GetStatus 获取订单状态
func (o *OrderListRequest) GetStatus() string {
	if o.Status == "" {
		return "已支付"
	}

	if o.Status == "全部" {
		return ""
	}

	return o.Status
}
