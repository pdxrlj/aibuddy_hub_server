// Package certs 证书管理
package certs

import (
	_ "embed"
)

// ApiclientCert 商户API证书内容
//
//go:embed apiclient_cert.pem
var ApiclientCert string

// ApiclientKey 商户API私钥内容
//
//go:embed apiclient_key.pem
var ApiclientKey string

// WechatpayPublicKey 微信支付公钥内容
//
//go:embed wechatpay_public_key_path.pem
var WechatpayPublicKey string
