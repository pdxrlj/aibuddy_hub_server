package pay

import (
	"context"
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/wechat/v3"
	"github.com/pkg/errors"
)

var _ IPay = (*WxMinPay)(nil)

// DefaultExpireTime 默认订单过期时间
const DefaultExpireTime = 10 * time.Minute

// WxMinPay 微信小程序支付
type WxMinPay struct {
	AppID              string
	MchID              string
	APIV3Key           string
	PrivateKey         string
	SerialNo           string
	WxPublicKeyContent []byte
	WxPublicKey        string
	Debug              bool
	OrderNotifyURL     string
	RefundNotifyURL    string
	*wechat.ClientV3
}

// WxMinPayOptions 微信小程序支付选项函数
type WxMinPayOptions func(*WxMinPay)

// WithAppID 设置AppID
func WithAppID(appID string) WxMinPayOptions {
	return func(pay *WxMinPay) {
		if appID != "" {
			pay.AppID = appID
		}
	}
}

// WithDebug 设置调试模式
func WithDebug(debug bool) WxMinPayOptions {
	return func(pay *WxMinPay) {
		pay.Debug = debug
	}
}

// WithMchID 设置商户号
func WithMchID(mchID string) WxMinPayOptions {
	return func(pay *WxMinPay) {
		if mchID != "" {
			pay.MchID = mchID
		}
	}
}

// WithSerialNo 设置证书序列号
func WithSerialNo(serialNo string) WxMinPayOptions {
	return func(pay *WxMinPay) {
		if serialNo != "" {
			pay.SerialNo = serialNo
		}
	}
}

// WithAPIV3Key 设置API V3密钥
func WithAPIV3Key(apiV3Key string) WxMinPayOptions {
	return func(pay *WxMinPay) {
		if apiV3Key != "" {
			pay.APIV3Key = apiV3Key
		}
	}
}

// WithPrivateKey 设置商户私钥
func WithPrivateKey(privateKey string) WxMinPayOptions {
	return func(pay *WxMinPay) {
		if privateKey != "" {
			pay.PrivateKey = privateKey
		}
	}
}

// WithWxPublicKeyContent 设置微信支付公钥内容
func WithWxPublicKeyContent(wxPublicKeyContent []byte) WxMinPayOptions {
	return func(pay *WxMinPay) {
		if wxPublicKeyContent != nil {
			pay.WxPublicKeyContent = wxPublicKeyContent
		}
	}
}

// WithWxPublicKey 设置微信支付公钥ID
func WithWxPublicKey(wxPublicKey string) WxMinPayOptions {
	return func(pay *WxMinPay) {
		if wxPublicKey != "" {
			pay.WxPublicKey = wxPublicKey
		}
	}
}

// WithOrderNotifyURL 设置订单通知URL
func WithOrderNotifyURL(orderNotifyURL string) WxMinPayOptions {
	return func(pay *WxMinPay) {
		if orderNotifyURL != "" {
			pay.OrderNotifyURL = orderNotifyURL
		}
	}
}

// WithRefundNotifyURL 设置退款通知URL
func WithRefundNotifyURL(refundNotifyURL string) WxMinPayOptions {
	return func(pay *WxMinPay) {
		if refundNotifyURL != "" {
			pay.RefundNotifyURL = refundNotifyURL
		}
	}
}

// NewWxMinPay 创建微信小程序支付实例
func NewWxMinPay(options ...WxMinPayOptions) (*WxMinPay, error) {
	pay := &WxMinPay{}
	for _, option := range options {
		option(pay)
	}

	if err := checkOptions(pay); err != nil {
		return nil, err
	}

	client, err := wechat.NewClientV3(pay.MchID, pay.SerialNo, pay.APIV3Key, pay.PrivateKey)
	if err != nil {
		return nil, err
	}

	if err := client.AutoVerifySignByPublicKey(pay.WxPublicKeyContent, pay.WxPublicKey); err != nil {
		return nil, err
	}
	pay.ClientV3 = client

	if pay.Debug {
		pay.DebugSwitch = gopay.DebugOn
	}

	return pay, nil
}

// WxPublicKeyMap 获取微信公钥映射
func (p *WxMinPay) WxPublicKeyMap() map[string]*rsa.PublicKey {
	return p.ClientV3.WxPublicKeyMap()
}

func checkOptions(pay *WxMinPay) error {
	if pay.AppID == "" {
		return ErrAppIDEmpty
	}
	if pay.MchID == "" {
		return ErrMchIDEmpty
	}
	if pay.APIV3Key == "" {
		return ErrWxAPIV3KeyEmpty
	}
	if pay.PrivateKey == "" {
		return ErrWxPrivateKeyEmpty
	}
	if pay.SerialNo == "" {
		return ErrSerialNoEmpty
	}
	if pay.WxPublicKey == "" {
		return ErrWxPublicKeyEmpty
	}
	if pay.WxPublicKeyContent == nil {
		return ErrWxPublicKeyContentEmpty
	}
	return nil
}

// WithWxDescription 设置订单描述
func WithWxDescription(description string) RequestOptions {
	return func(req *CreateOrderRequest) {
		req.Description = description
	}
}

// WithWxOutTradeNo 设置商户订单号
func WithWxOutTradeNo(outTradeNo string) RequestOptions {
	return func(req *CreateOrderRequest) {
		req.OutTradeNo = outTradeNo
	}
}

// WithWxTotalAmount 设置订单金额
func WithWxTotalAmount(totalAmount int64) RequestOptions {
	return func(req *CreateOrderRequest) {
		req.TotalAmount = totalAmount
	}
}

// WithWxCurrency 设置货币类型
func WithWxCurrency(currency string) RequestOptions {
	return func(req *CreateOrderRequest) {
		req.Currency = currency
	}
}

// WithWxPayerOpenID 设置支付者OpenID
func WithWxPayerOpenID(openID string) RequestOptions {
	return func(req *CreateOrderRequest) {
		req.Payer.OpenID = openID
	}
}

// WithWxExtraFields 设置额外字段
func WithWxExtraFields(extraFields map[string]any) RequestOptions {
	return func(req *CreateOrderRequest) {
		req.ExtraFields = extraFields
	}
}

// DefaultWxOrderRequest 创建默认微信订单请求
func DefaultWxOrderRequest() *CreateOrderRequest {
	return &CreateOrderRequest{
		Currency: "CNY",
	}
}

// CreateOrder 创建订单
func (p *WxMinPay) CreateOrder(ctx context.Context, req *CreateOrderRequest, options ...RequestOptions) (any, error) {
	for _, option := range options {
		option(req)
	}

	expireTime := time.Now().Add(DefaultExpireTime).Format(time.RFC3339)
	bm := make(gopay.BodyMap)
	bm.Set("appid", p.AppID).
		Set("mchid", p.MchID).
		Set("description", req.Description).
		Set("out_trade_no", req.OutTradeNo).
		Set("time_expire", expireTime).
		Set("notify_url", p.OrderNotifyURL).
		SetBodyMap("amount", func(bm gopay.BodyMap) {
			bm.Set("total", req.TotalAmount).
				Set("currency", req.Currency)
		}).
		SetBodyMap("payer", func(bm gopay.BodyMap) {
			bm.Set("openid", req.Payer.OpenID)
		})
	if req.ExtraFields != nil {
		for k, v := range req.ExtraFields {
			bm.Set(k, v)
		}
	}

	wxRsp, err := p.V3TransactionJsapi(ctx, bm)
	if err != nil {
		return nil, err
	}
	if wxRsp.Code == wechat.Success {
		return wxRsp, nil
	}

	return nil, errors.Wrap(ErrWxCreateOrderFailed, fmt.Sprintf(" code: %d, message: %s", wxRsp.Code, wxRsp.ErrResponse.Message))
}

// RefundOrder 退款
func (p *WxMinPay) RefundOrder(ctx context.Context, req *RefundOrderRequest) (any, error) {
	bm := make(gopay.BodyMap)
	bm.Set("out_trade_no", req.OutTradeNo).
		Set("out_refund_no", req.OutRefundNo).
		SetBodyMap("amount", func(bm gopay.BodyMap) {
			bm.Set("refund", req.TotalAmount).
				Set("currency", req.Currency).
				Set("total", req.OriginalAmount)
		}).
		Set("notify_url", p.RefundNotifyURL)

	if req.Reason != "" {
		bm.Set("reason", req.Reason)
	}

	wxRsp, err := p.V3Refund(ctx, bm)
	if err != nil {
		return nil, err
	}
	if wxRsp.Code == wechat.Success {
		return wxRsp, nil
	}

	return nil, errors.Wrap(ErrWxRefundOrderFailed, fmt.Sprintf(" code: %d, message: %s", wxRsp.Code, wxRsp.ErrResponse.Message))
}

// GetOrderStatus 查询订单状态
func (p *WxMinPay) GetOrderStatus(ctx context.Context, outTradeNo string) (any, error) {
	wxRsp, err := p.V3TransactionQueryOrder(ctx, wechat.OutTradeNo, outTradeNo)
	if err != nil {
		return nil, err
	}
	if wxRsp.Code == wechat.Success {
		return wxRsp, nil
	}

	return nil, errors.Wrap(ErrWxGetOrderStatusFailed, fmt.Sprintf(" code: %d, message: %s", wxRsp.Code, wxRsp.ErrResponse.Message))
}

// CloseOrder 关闭订单
func (p *WxMinPay) CloseOrder(ctx context.Context, outTradeNo string) (any, error) {
	wxRsp, err := p.V3TransactionCloseOrder(ctx, outTradeNo)
	if err != nil {
		return nil, err
	}
	if wxRsp.Code == wechat.Success {
		return wxRsp, nil
	}
	return nil, errors.Wrap(ErrWxCloseOrderFailed, fmt.Sprintf(" code: %d, message: %s", wxRsp.Code, wxRsp.ErrResponse.Message))
}

// PaySignOfJSAPI 获取JSAPI支付签名
func (p *WxMinPay) PaySignOfJSAPI(appid, prepayid string) (any, error) {
	sign, err := p.ClientV3.PaySignOfJSAPI(appid, prepayid)
	if err != nil {
		return nil, err
	}
	return sign, nil
}
