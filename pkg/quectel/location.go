// Package quec 提供移远物联卡 API 客户端功能
package quec

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// LocationResponse 设备位置查询响应
type LocationResponse struct {
	ResultCode    int          `json:"resultCode"`
	ErrorMessage  string       `json:"errorMessage"`
	Data          LocationData `json:"data"`
}

// LocationData 位置数据
type LocationData struct {
	Msisdn         string `json:"MSISDN"`
	Iccid          string `json:"ICCID"`
	Latitude       string `json:"LATITUDE"`
	Longitude      string `json:"LONGITUDE"`
	PositionResult string `json:"POSITIONRESULT,omitempty"`
	MsidType       string `json:"MSID_TYPE,omitempty"`
	Msid           string `json:"MSID,omitempty"`
	LocalTime      string `json:"LOCALTIME,omitempty"`
}

// GetDeviceLocation 查询设备位置
func GetDeviceLocation(appKey, timestamp, secretKey string, iccid string) (*LocationResponse, error) {
	params := map[string]string{
		"appKey": appKey,
		"method": "fc.function.card.location",
		"t":      timestamp,
	}
	if iccid != "" {
		params["iccid"] = iccid
	}
	sign := Sign(params, secretKey)
	params["sign"] = sign

	fmt.Printf("Request params: %v\n", params)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	formData := url.Values{}
	for k, v := range params {
		formData.Set(k, v)
	}

	resp, err := client.PostForm("https://api.quectel.com/openapi/router", formData)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	fmt.Printf("Response: %s\n", string(body))

	var result LocationResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response failed, raw: %s, error: %w", string(body), err)
	}

	return &result, nil
}
