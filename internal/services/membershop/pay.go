// Package membershop 提供会员商城相关功能
package membershop

import (
	"aibuddy/pkg/config"
	"aibuddy/pkg/pay"
	"encoding/json"
	"net/http"

	"github.com/go-pay/gopay/wechat/v3"
)

// PayNotifyData 支付通知数据
type PayNotifyData struct {
	*wechat.V3DecryptPayResult
	notify *wechat.V3NotifyReq
}

// NewPayNotifyData 创建支付通知数据
func NewPayNotifyData(notify *wechat.V3NotifyReq) *PayNotifyData {
	return &PayNotifyData{
		V3DecryptPayResult: &wechat.V3DecryptPayResult{},
		notify:             notify,
	}
}

// Decrypt 解密
func (s *PayNotifyData) Decrypt() error {
	PayConfig := config.Instance.Pay

	if err := wechat.V3DecryptNotifyCipherTextToStruct(
		s.notify.Resource.Ciphertext,
		s.notify.Resource.Nonce,
		s.notify.Resource.AssociatedData,
		PayConfig.APIV3Key,
		s.V3DecryptPayResult,
	); err != nil {
		return err
	}
	return nil
}

// VerifySign 验签
func (s *PayNotifyData) VerifySign(pay pay.IPay) error {
	if err := s.notify.VerifySignByPKMap(pay.WxPublicKeyMap()); err != nil {
		return err
	}
	return nil
}

// PayRefundNotifyData 退款通知数据
type PayRefundNotifyData struct {
	*wechat.V3DecryptRefundResult
	notify *wechat.V3NotifyReq
}

// NewPayRefundNotifyData 创建退款通知数据
func NewPayRefundNotifyData(notify *wechat.V3NotifyReq) *PayRefundNotifyData {
	return &PayRefundNotifyData{
		V3DecryptRefundResult: &wechat.V3DecryptRefundResult{},
		notify:                notify,
	}
}

// NewShoppingRefundNotifyData 创建退款通知数据
func NewShoppingRefundNotifyData(notify *wechat.V3NotifyReq) *PayRefundNotifyData {
	return &PayRefundNotifyData{
		V3DecryptRefundResult: &wechat.V3DecryptRefundResult{},
		notify:                notify,
	}
}

// Decrypt 解密
func (s *PayRefundNotifyData) Decrypt() error {
	PayConfig := config.Instance.Pay

	if err := wechat.V3DecryptNotifyCipherTextToStruct(
		s.notify.Resource.Ciphertext,
		s.notify.Resource.Nonce,
		s.notify.Resource.AssociatedData,
		PayConfig.APIV3Key,
		s.V3DecryptRefundResult,
	); err != nil {
		return err
	}
	return nil
}

// VerifySign 验签
func (s *PayRefundNotifyData) VerifySign(pay pay.IPay) error {
	if err := s.notify.VerifySignByPKMap(pay.WxPublicKeyMap()); err != nil {
		return err
	}
	return nil
}

// PayResponse 返回微信支付回调响应
func PayResponse(w http.ResponseWriter, code string, message string) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	
	rsp := &wechat.V3NotifyRsp{
		Code:    code,
		Message: message,
	}
	
	return json.NewEncoder(w).Encode(rsp)
}
