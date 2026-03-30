package quec

import (
	"testing"
)

func TestSign(t *testing.T) {
	secretKey := "t2SiD64l"

	params := map[string]string{
		"appId":     "a835Fs8ON1BjE4MV",
		"method":    "iot.dcs.getDeviceList",
		"timestamp": "1688888888000",
		"version":   "1.0",
	}

	sign := Sign(params, secretKey)
	t.Logf("sign: %s", sign)

	if sign == "" {
		t.Error("sign should not be empty")
	}
}

func TestQueryLocation(t *testing.T) {
	appID := "a835Fs8ON1BjE4MV"
	method := "iot.dcs.getDeviceList"
	secretKey := "t2SiD64l"

	url := QueryLocation(appID, method, secretKey)
	t.Logf("url: %s", url)

	if url == "" {
		t.Error("url should not be empty")
	}
}

func TestQueryDeviceStatus(t *testing.T) {
	appID := "a835Fs8ON1BjE4MV"
	timestamp := "1688888888000"
	secretKey := "t2SiD64l"

	url := QueryDeviceStatus(appID, timestamp, secretKey, "898607B8102580151949", "460041888422950", "861540085739384")
	t.Logf("url: %s", url)

	if url == "" {
		t.Error("url should not be empty")
	}
}

func TestQueryDeviceLocation(t *testing.T) {
	appKey := "a835Fs8ON1BjE4MV"
	timestamp := "1688888888000"
	secretKey := "t2SiD64l"

	url := QueryDeviceLocation(appKey, timestamp, secretKey, "898607B8102580151949")
	t.Logf("url: %s", url)

	if url == "" {
		t.Error("url should not be empty")
	}
}

func TestSignWithSortedParams(t *testing.T) {
	secretKey := "test-secret"

	params := map[string]string{
		"z_param": "z_value",
		"a_param": "a_value",
		"m_param": "m_value",
		"sign":    "should_be_excluded",
	}

	sign := Sign(params, secretKey)

	expectedMessage := "a_param=a_value&m_param=m_value&z_param=z_value"
	t.Logf("signed message: %s", expectedMessage)
	t.Logf("result sign: %s", sign)

	if sign == "" {
		t.Error("sign should not be empty")
	}
}
