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

// MonthFlowResponse 月流量查询响应
type MonthFlowResponse struct {
	ResultCode   int     `json:"resultCode"`
	ErrorMessage string  `json:"errorMessage"`
	MonthFlow    float64 `json:"monthflow"`  // 指定月份的当月用量(MB)
	CycleFlow    float64 `json:"cycleflow"`  // 计费周期内，截至指定月份的已用总流量(MB)
}

// BatchMonthFlowResponse 批量月流量查询响应
type BatchMonthFlowResponse struct {
	ResultCode   int                    `json:"resultCode"`
	ErrorMessage string                 `json:"errorMessage"`
	Data         []BatchMonthFlowItem   `json:"data"`
}

// BatchMonthFlowItem 批量月流量查询项
type BatchMonthFlowItem struct {
	Number    string  `json:"number"`    // 默认为MSISDN，如果传入ICCID参数查询，则返回ICCID
	MonthFlow  float64 `json:"monthflow"` // 指定月份的当月用量(MB)
	CycleFlow  float64 `json:"cycleflow"` // 计费周期内，截至指定月份的已用总流量(MB)
}

// MonthFlowListResponse 资产月流量日志响应
type MonthFlowListResponse struct {
	ResultCode   int              `json:"resultCode"`
	ErrorMessage string           `json:"errorMessage"`
	Page         PageInfo         `json:"page"`
	Flows        []MonthFlowItem  `json:"flows"`
}

// PageInfo 分页信息
type PageInfo struct {
	Page       int `json:"page"`       // 分页页码
	PageSize   int `json:"pageSize"`   // 分页大小
	TotalCount int `json:"totalCount"` // 记录总数
	TotalPage  int `json:"totalPage"`  // 总页数
}

// MonthFlowItem 月流量项
type MonthFlowItem struct {
	Iccid     string  `json:"iccid"`     // ICCID
	Msisdn    string  `json:"msisdn"`    // MSISDN
	MonthFlow float64 `json:"monthflow"` // 当月用量(MB)
	Flow      float64 `json:"flow"`      // 计费周期内已用总流量(MB)
}

// DayFlowResponse 资产日用量日志响应
type DayFlowResponse struct {
	ResultCode   int           `json:"resultCode"`
	ErrorMessage string        `json:"errorMessage"`
	Flows        []DayFlowItem `json:"flows"`
}

// DayFlowItem 日用量项
type DayFlowItem struct {
	Date string  `json:"date"` // 日期，格式：yyyyMMdd
	Flow float64 `json:"flow"` // 日流量(MB)
}

// GetMonthFlow 查询单张物联卡在指定月份的月用量数据
// month: 指定月份，格式：yyyyMM (如：202403)
// iccid: 集成电路卡识别码
func GetMonthFlow(appKey, secretKey, month, iccid string) (*MonthFlowResponse, error) {
	params := map[string]string{
		"appKey": appKey,
		"method": "fc.function.card.monthflow",
		"t":      fmt.Sprintf("%d", time.Now().Unix()),
		"month":  month,
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

	var result MonthFlowResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response failed, raw: %s, error: %w", string(body), err)
	}

	return &result, nil
}

// GetBatchMonthFlow 批量查询物联卡在指定月份的月用量数据
// month: 指定月份，格式：yyyyMM
// iccids: 多个ICCID，以英文逗号隔开，最多1000张
func GetBatchMonthFlow(appKey, secretKey, month, iccids string) (*BatchMonthFlowResponse, error) {
	params := map[string]string{
		"appKey": appKey,
		"method": "fc.function.cards.monthflow",
		"t":      fmt.Sprintf("%d", time.Now().Unix()),
		"month":  month,
	}
	if iccids != "" {
		params["iccids"] = iccids
	}
	sign := Sign(params, secretKey)
	params["sign"] = sign

	return doRequest[BatchMonthFlowResponse](params)
}

// GetMonthFlowList 获取用户资产指定月的月流量日志
// month: 指定月份，格式：yyyyMM
// pageNo: 分页页码，默认为1
// pageSize: 分页大小，默认为100，最大1000
func GetMonthFlowList(appKey, secretKey, month string, pageNo, pageSize int) (*MonthFlowListResponse, error) {
	params := map[string]string{
		"appKey": appKey,
		"method": "fc.function.monthflow.list",
		"t":      fmt.Sprintf("%d", time.Now().Unix()),
		"month":  month,
	}
	if pageNo > 0 {
		params["pageNo"] = fmt.Sprintf("%d", pageNo)
	}
	if pageSize > 0 {
		params["pageSize"] = fmt.Sprintf("%d", pageSize)
	}
	sign := Sign(params, secretKey)
	params["sign"] = sign

	return doRequest[MonthFlowListResponse](params)
}

// GetDayFlowList 资产当月或指定月份用量日志查询
// month: 指定月份，格式：yyyyMM
// iccid: 集成电路卡识别码
func GetDayFlowList(appKey, secretKey, month, iccid string) (*DayFlowResponse, error) {
	params := map[string]string{
		"appKey": appKey,
		"method": "fc.function.dayflow.list",
		"t":      fmt.Sprintf("%d", time.Now().Unix()),
		"month":  month,
	}
	if iccid != "" {
		params["iccid"] = iccid
	}
	sign := Sign(params, secretKey)
	params["sign"] = sign

	return doRequest[DayFlowResponse](params)
}

// doRequest 通用请求方法
func doRequest[T any](params map[string]string) (*T, error) {
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

	var result T
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response failed, raw: %s, error: %w", string(body), err)
	}

	return &result, nil
}
