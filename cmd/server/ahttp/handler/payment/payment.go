// Package payment 提供支付回调处理
package payment

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"aibuddy/pkg/ahttp"
	"aibuddy/pkg/config"
	"aibuddy/pkg/pay"
	"aibuddy/pkg/pay/certs"

	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/wechat/v3"
)

// Handler 支付回调处理器
type Handler struct {
	payInstance pay.IPay
	apiV3Key    string
}

// NewHandler 创建支付回调处理器
func NewHandler() *Handler {
	cfg := config.Instance
	payInstance, err := pay.Instance(&pay.Config{
		AppID:              cfg.Pay.AppID,
		MchID:              cfg.Pay.MchID,
		APIV3Key:           cfg.Pay.APIV3Key,
		SerialNo:           cfg.Pay.SerialNo,
		WxPublicKey:        cfg.Pay.WechatpaySerialNo,
		OrderNotifyURL:     cfg.Pay.NotifyURL,
		RefundNotifyURL:    cfg.Pay.RefundNotifyURL,
		PrivateKey:         certs.ApiclientKey,
		WxPublicKeyContent: []byte(certs.WechatpayPublicKey),
	})
	if err != nil {
		panic(fmt.Sprintf("初始化支付实例失败: %v", err))
	}

	return &Handler{
		payInstance: payInstance,
		apiV3Key:    cfg.Pay.APIV3Key,
	}
}

// writeResponse 写入HTTP响应
func writeResponse(w http.ResponseWriter, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := &wechat.V3NotifyRsp{
		Code:    code,
		Message: message,
	}
	data, _ := json.Marshal(resp)
	_, _ = w.Write(data)
}

// PayNotify 支付结果通知回调
func (h *Handler) PayNotify(state *ahttp.State) error {
	return h.handleNotify(state, newPayNotifyHandler, "PayNotify")
}

// RefundNotify 退款结果通知回调
func (h *Handler) RefundNotify(state *ahttp.State) error {
	return h.handleNotify(state, newRefundNotifyHandler, "RefundNotify")
}

// notifyHandlerFactory 通知处理策略工厂函数类型
type notifyHandlerFactory func(notify *wechat.V3NotifyReq) notifyHandler

// handleNotify 处理微信支付/退款通知
func (h *Handler) handleNotify(state *ahttp.State, factory notifyHandlerFactory, tag string) error {
	w := state.Ctx.Response()
	req := state.Ctx.Request()

	slog.Info(fmt.Sprintf("[%s] 开始处理通知", tag))

	fail := func(msg string, err error) {
		slog.Error(fmt.Sprintf("[%s] %s", tag, msg), "error", err)
		writeResponse(w, gopay.FAIL, msg)
	}

	notifyReq, err := wechat.V3ParseNotify(req)
	if err != nil {
		fail("解析通知失败", err)
		return nil
	}

	notifyJSON, _ := json.MarshalIndent(notifyReq, "", "  ")
	slog.Info(fmt.Sprintf("[%s] 原始通知数据", tag), "data", string(notifyJSON))

	if err := notifyReq.VerifySignByPKMap(h.payInstance.WxPublicKeyMap()); err != nil {
		fail("验证签名失败", err)
		return nil
	}

	handler := factory(notifyReq)
	resultJSON, err := handler.Decrypt(h.apiV3Key)
	if err != nil {
		fail("解密通知数据失败", err)
		return nil
	}

	slog.Info(fmt.Sprintf("[%s] 解密数据", tag), "data", string(resultJSON))
	slog.Info(fmt.Sprintf("[%s] 通知处理完成", tag))
	writeResponse(w, gopay.SUCCESS, "成功")
	return nil
}
