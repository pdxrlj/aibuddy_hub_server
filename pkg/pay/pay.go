// Package pay 提供微信支付相关功能
package pay

import (
	"context"
	"crypto/rsa"
)

// Type 支付类型
type Type string

// 支付类型常量
const (
	// TypeWxMin 微信小程序支付
	TypeWxMin Type = "wx_min"
)

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	Description string                  `json:"description"`
	OutTradeNo  string                  `json:"out_trade_no"`
	TimeExpire  string                  `json:"time_expire"`
	TotalAmount int64                   `json:"total_amount"`
	Currency    string                  `json:"currency"`
	Payer       CreateOrderRequestPayer `json:"payer"`

	ExtraFields map[string]any `json:"extra_fields,omitempty"`
}

// CreateOrderRequestPayer 支付者信息
type CreateOrderRequestPayer struct {
	OpenID string `json:"openid"`
}

// RefundOrderRequest 退款请求
type RefundOrderRequest struct {
	OutTradeNo     string `json:"out_trade_no"`     // 原订单号
	OutRefundNo    string `json:"out_refund_no"`    // 退款订单号
	TotalAmount    int64  `json:"total_amount"`     // 退款金额，单位分
	OriginalAmount int64  `json:"original_amount"`  // 原订单总金额，单位分
	Currency       string `json:"currency"`         // 退款金额币种，CNY
	Reason         string `json:"reason,omitempty"` // 退款原因
}

// GetOrderStatusRequest 查询订单状态请求
type GetOrderStatusRequest struct {
	OutTradeNo string `json:"out_trade_no"` // 原订单号
}

// RequestOptions 支付请求选项函数
type RequestOptions func(req *CreateOrderRequest)

// IPay 支付接口
type IPay interface {
	CreateOrder(ctx context.Context, req *CreateOrderRequest, options ...RequestOptions) (any, error)
	RefundOrder(ctx context.Context, req *RefundOrderRequest) (any, error)
	GetOrderStatus(ctx context.Context, outTradeNo string) (any, error)
	CloseOrder(ctx context.Context, outTradeNo string) (any, error)
	PaySignOfJSAPI(appid, prepayid string) (any, error)
	WxPublicKeyMap() map[string]*rsa.PublicKey
}

// Config 支付配置
type Config struct {
	AppID             string
	MchID             string
	APIV3Key          string
	PrivateKey        string
	SerialNo          string
	WxPublicKeyContent []byte
	WxPublicKey       string
	Debug             bool `json:"-"`
	OrderNotifyURL    string
	RefundNotifyURL   string
}

// WxPayDefault 创建默认微信支付实例
func WxPayDefault(config *Config) (IPay, error) {
	pay, err := NewWxMinPay(
		WithAppID(config.AppID),
		WithDebug(config.Debug),
		WithMchID(config.MchID),
		WithAPIV3Key(config.APIV3Key),
		WithPrivateKey(config.PrivateKey),
		WithSerialNo(config.SerialNo),
		WithWxPublicKeyContent(config.WxPublicKeyContent),
		WithWxPublicKey(config.WxPublicKey),
		WithOrderNotifyURL(config.OrderNotifyURL),
		WithRefundNotifyURL(config.RefundNotifyURL),
	)
	if err != nil {
		return nil, err
	}
	return pay, nil
}

// Instance 创建支付实例
func Instance(config *Config, payType ...Type) (IPay, error) {
	// TODO 支持其他支付类型
	if len(payType) == 0 || payType[0] == TypeWxMin {
		return WxPayDefault(config)
	}
	return nil, ErrTypeNotSupported
}
