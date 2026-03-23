// Package pay 提供微信支付相关功能
package pay

import "errors"

// 支付相关错误定义
var (
	// ErrAppIDEmpty AppID为空错误
	ErrAppIDEmpty = errors.New("appid 错误")
	// ErrMchIDEmpty 商户号为空错误
	ErrMchIDEmpty = errors.New("mchid 错误")

	// ErrWxAPIV3KeyEmpty API V3密钥为空错误
	ErrWxAPIV3KeyEmpty = errors.New("APIV3Key 错误")
	// ErrWxPrivateKeyEmpty 私钥为空错误
	ErrWxPrivateKeyEmpty = errors.New("PrivateKey 错误")
	// ErrSerialNoEmpty 证书序列号为空错误
	ErrSerialNoEmpty = errors.New("SerialNo 错误")

	// ErrWxPublicKeyEmpty 微信公钥ID为空错误
	ErrWxPublicKeyEmpty = errors.New("WxPublicKey 错误")
	// ErrWxPublicKeyContentEmpty 微信公钥内容为空错误
	ErrWxPublicKeyContentEmpty = errors.New("WxPublicKeyContent 错误")

	// ErrWxOrderNotifyURLEmpty 订单通知URL为空错误
	ErrWxOrderNotifyURLEmpty = errors.New("WxOrderNotifyURL 错误")
	// ErrWxRefundNotifyURLEmpty 退款通知URL为空错误
	ErrWxRefundNotifyURLEmpty = errors.New("WxRefundNotifyURL 错误")

	// ErrWxCreateOrderFailed 创建订单失败错误
	ErrWxCreateOrderFailed = errors.New("CreateOrder failed")
	// ErrWxRefundOrderFailed 退款失败错误
	ErrWxRefundOrderFailed = errors.New("RefundOrder failed")
	// ErrWxGetOrderStatusFailed 查询订单状态失败错误
	ErrWxGetOrderStatusFailed = errors.New("GetOrderStatus failed")

	// ErrTypeNotSupported 支付类型不支持错误
	ErrTypeNotSupported = errors.New("PayType not supported")

	// ErrWxCloseOrderFailed 关闭订单失败错误
	ErrWxCloseOrderFailed = errors.New("CloseOrder failed")
)
