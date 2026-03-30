package quec

import (
	"testing"
)

func TestSignExample(t *testing.T) {
	// 文档示例
	appKey := "000001"
	method := "fc.function.card.info"
	msisdn := "1064868950247"
	timestamp := "1732605510"
	secret := "abcdef"

	params := map[string]string{
		"appKey": appKey,
		"method": method,
		"msisdn": msisdn,
		"t":      timestamp,
	}

	sign := Sign(params, secret)
	t.Logf("Generated sign: %s", sign)
	t.Logf("Expected sign: 7AAD18B95E908CB8F80AA3518A7E65426D4C7BBB")

	if sign == "7AAD18B95E908CB8F80AA3518A7E65426D4C7BBB" || sign == "7aad18b95e908cb8f80aa3518a7e65426d4c7bbb" {
		t.Log("签名验证通过!")
	} else {
		t.Error("签名验证失败")
	}
}
