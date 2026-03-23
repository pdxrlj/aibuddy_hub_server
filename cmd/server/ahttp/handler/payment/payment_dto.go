// Package payment 提供支付回调处理
package payment

import (
	"encoding/json"

	"github.com/go-pay/gopay/wechat/v3"
)

// notifyHandler 通知处理策略接口
type notifyHandler interface {
	// Decrypt 解密通知数据并返回结构化的 JSON
	Decrypt(apiV3Key string) ([]byte, error)
}

// payNotifyHandler 支付通知处理策略
type payNotifyHandler struct {
	notify *wechat.V3NotifyReq
	result *wechat.V3DecryptPayResult
}

func newPayNotifyHandler(notify *wechat.V3NotifyReq) notifyHandler {
	return &payNotifyHandler{
		notify: notify,
		result: &wechat.V3DecryptPayResult{},
	}
}

func (h *payNotifyHandler) Decrypt(apiV3Key string) ([]byte, error) {
	err := wechat.V3DecryptNotifyCipherTextToStruct(
		h.notify.Resource.Ciphertext,
		h.notify.Resource.Nonce,
		h.notify.Resource.AssociatedData,
		apiV3Key,
		h.result,
	)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(h.result, "", "  ")
}

// refundNotifyHandler 退款通知处理策略
type refundNotifyHandler struct {
	notify *wechat.V3NotifyReq
	result *wechat.V3DecryptRefundResult
}

func newRefundNotifyHandler(notify *wechat.V3NotifyReq) notifyHandler {
	return &refundNotifyHandler{
		notify: notify,
		result: &wechat.V3DecryptRefundResult{},
	}
}

func (h *refundNotifyHandler) Decrypt(apiV3Key string) ([]byte, error) {
	err := wechat.V3DecryptNotifyCipherTextToStruct(
		h.notify.Resource.Ciphertext,
		h.notify.Resource.Nonce,
		h.notify.Resource.AssociatedData,
		apiV3Key,
		h.result,
	)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(h.result, "", "  ")
}
