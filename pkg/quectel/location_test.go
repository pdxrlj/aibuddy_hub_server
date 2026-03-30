package quec

import (
	"fmt"
	"testing"
	"time"
)

func TestGetDeviceLocation(t *testing.T) {
	appKey := "a835Fs8ON1BjE4MV"
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	secretKey := "t2SiD64l"

	result, err := GetDeviceLocation(appKey, timestamp, secretKey, "898607B8102580151949")
	if err != nil {
		t.Fatalf("GetDeviceLocation failed: %v", err)
	}

	t.Logf("ResultCode: %d, ErrorMessage: %s", result.ResultCode, result.ErrorMessage)

	if result.ResultCode != 0 {
		t.Logf("Response code: %d, ErrorMessage: %s", result.ResultCode, result.ErrorMessage)
		return
	}

	t.Logf("Location: lat=%s, lng=%s", result.Data.Latitude, result.Data.Longitude)
	t.Logf("MSISDN: %s, ICCID: %s", result.Data.Msisdn, result.Data.Iccid)
	t.Logf("PositionResult: %s", result.Data.PositionResult)
}
