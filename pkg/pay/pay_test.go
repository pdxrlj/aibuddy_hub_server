package pay

import (
	"aibuddy/pkg/config"
	"aibuddy/pkg/helpers"
	"aibuddy/pkg/pay/certs"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

var _Config *Config

func TestMain(m *testing.M) {
	cfg := config.Setup("../../config")

	_Config = &Config{
		AppID:              cfg.Pay.AppID,
		MchID:              cfg.Pay.MchID,
		APIV3Key:           cfg.Pay.APIV3Key,
		SerialNo:           cfg.Pay.SerialNo,
		WxPublicKey:        cfg.Pay.WechatpaySerialNo,
		OrderNotifyURL:     cfg.Pay.NotifyURL,
		RefundNotifyURL:    cfg.Pay.RefundNotifyURL,
		Debug:              true,
		PrivateKey:         certs.ApiclientKey,
		WxPublicKeyContent: []byte(certs.WechatpayPublicKey),
	}

	m.Run()
}

func TestInstance(t *testing.T) {
	pay, err := Instance(_Config)
	assert.NoError(t, err)
	assert.NotNil(t, pay)
}

func TestCreateOrder(t *testing.T) {
	pay, err := Instance(_Config)
	assert.NoError(t, err)
	assert.NotNil(t, pay)

	req := DefaultWxOrderRequest()
	req.Description = "测试订单"
	req.OutTradeNo = helpers.GenerateNumber(16)
	req.TotalAmount = 1
	req.Payer.OpenID = "o00vl187dORHD0SYq578_FpBCz2o"

	res, err := pay.CreateOrder(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	t.Logf("CreateOrder response: %+v", res)
}
