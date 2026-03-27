// Package quec 提供移远物联卡 API 客户端功能
package quec

import (
	"crypto/sha1"
	"fmt"
	"sort"
	"strings"
)

// Sign 生成签名
func Sign(params map[string]string, secretKey string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString(secretKey)
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString(params[k])
	}
	sb.WriteString(secretKey)

	h := sha1.New()
	h.Write([]byte(sb.String()))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// QueryLocation 构建位置查询URL
func QueryLocation(appID, method, secretKey string) string {
	params := map[string]string{
		"appKey": appID,
		"method": method,
	}
	sign := Sign(params, secretKey)
	params["sign"] = sign
	return buildRequestURL(params)
}

// QueryDeviceStatus 构建设备状态查询URL
func QueryDeviceStatus(appID, timestamp, secretKey string, iccid, imsi, imei string) string {
	params := map[string]string{
		"appKey": appID,
		"method": "iot.dcs.getDeviceStatus",
		"t":      timestamp,
	}
	if iccid != "" {
		params["iccid"] = iccid
	}
	if imsi != "" {
		params["imsi"] = imsi
	}
	if imei != "" {
		params["imei"] = imei
	}
	sign := Sign(params, secretKey)
	params["sign"] = sign
	return buildRequestURL(params)
}

// QueryDeviceLocation 构建设备位置查询URL
func QueryDeviceLocation(appKey, timestamp, secretKey string, iccid string) string {
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
	return buildRequestURL(params)
}

func buildRequestURL(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString("https://api.quectel.com/openapi/router?")
	for i, k := range keys {
		if i > 0 {
			sb.WriteString("&")
		}
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(params[k])
	}
	return sb.String()
}
