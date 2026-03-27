// Package baidu 百度地图硬件定位API
package baidu

import (
	"aibuddy/pkg/config"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	// MapsBaseURL 百度地图API基础URL
	MapsBaseURL = "https://api.map.baidu.com"
)

// IMSIInfo IMSI解析结果
type IMSIInfo struct {
	MCC int    // 移动国家代码 (Mobile Country Code)
	MNC int    // 移动网络代码 (Mobile Network Code)
	MSIN string // 移动用户识别码 (Mobile Subscription Identification Number)
}

// ParseIMSI 解析IMSI获取MCC和MNC
// IMSI格式: MCC(3位) + MNC(2位) + MSIN(10位)
// 例如: 460041888422950 -> MCC=460, MNC=04, MSIN=1888422950
func ParseIMSI(imsi string) (*IMSIInfo, error) {
	if len(imsi) != 15 {
		return nil, fmt.Errorf("IMSI长度必须为15位，当前: %d", len(imsi))
	}

	mcc, err := strconv.Atoi(imsi[:3])
	if err != nil {
		return nil, fmt.Errorf("解析MCC失败: %w", err)
	}

	mnc, err := strconv.Atoi(imsi[3:5])
	if err != nil {
		return nil, fmt.Errorf("解析MNC失败: %w", err)
	}

	return &IMSIInfo{
		MCC:  mcc,
		MNC:  mnc,
		MSIN: imsi[5:],
	}, nil
}

// MapsClient 百度地图客户端
type MapsClient struct {
	httpClient *resty.Client
	ak         string
	appID      string
}

// NewMapsClient 创建百度地图客户端
func NewMapsClient() *MapsClient {
	cfg := config.Instance.Baidu.Maps
	return &MapsClient{
		httpClient: resty.New().
			SetRetryCount(3).
			SetRetryWaitTime(500 * time.Millisecond).
			SetRetryMaxWaitTime(5 * time.Second).
			AddRetryCondition(func(r *resty.Response, err error) bool {
				if err != nil {
					return true
				}
				return r.StatusCode() >= http.StatusInternalServerError
			}),
		ak:    cfg.Ak,
		appID: cfg.AppID,
	}
}

// NewMapsClientWithAK 使用指定AK创建百度地图客户端
func NewMapsClientWithAK(ak, appID string) *MapsClient {
	return &MapsClient{
		httpClient: resty.New().
			SetRetryCount(3).
			SetRetryWaitTime(500 * time.Millisecond).
			SetRetryMaxWaitTime(5 * time.Second).
			AddRetryCondition(func(r *resty.Response, err error) bool {
				if err != nil {
					return true
				}
				return r.StatusCode() >= http.StatusInternalServerError
			}),
		ak:    ak,
		appID: appID,
	}
}

// WiFiAccessPoint WiFi接入点信息
type WiFiAccessPoint struct {
	MAC    string `json:"mac"`    // WiFi MAC地址，格式: 54:6c:3e:77:89:ab
	RSSI   int    `json:"rssi"`   // 信号强度，负值，如 -45
	SSID   string `json:"ssid"`   // WiFi名称，可选
	Freq   int    `json:"freq"`   // 频率(MHz)，可选
	IsMain bool   `json:"isMain"` // 是否为主WiFi，可选
}

// BaseStation 基站信息
type BaseStation struct {
	MCC int     `json:"mcc"`           // 移动国家代码，中国为460
	MNC int     `json:"mnc"`           // 移动网络代码
	LAC int     `json:"lac"`           // 位置区域码
	CID int     `json:"cid"`           // 基站ID (Cell ID)
	Lat float64 `json:"lat,omitempty"` // 纬度，可选
	Lon float64 `json:"lon,omitempty"` // 经度，可选
}

// HardwareLocRequest 硬件定位请求
type HardwareLocRequest struct {
	// WiFi列表，WiFi定位时使用
	WiFiList []WiFiAccessPoint `json:"wifi_list,omitempty"`
	// 基站信息，基站定位时使用
	CellList []BaseStation `json:"cell_list,omitempty"`
	// 定位类型: wifi, cell, mixed
	LocType string `json:"loc_type,omitempty"`
	// 是否返回详细地址信息
	NeedAddress bool `json:"need_address,omitempty"`
	// 坐标类型: bd09ll(百度经纬度), gcj02(国测局), wgs84(GPS)
	CoordType string `json:"coord_type,omitempty"`
	// 设备IMEI，15位数字
	IMEI string `json:"imei,omitempty"`
	// 是否CDMA网络
	CDMA int `json:"cdma,omitempty"`
}

// HardwareLocResult 定位结果
type HardwareLocResult struct {
	Lat         float64 `json:"lat"`          // 纬度
	Lon         float64 `json:"lon"`          // 经度
	Radius      float64 `json:"radius"`       // 定位精度半径(米)
	Confidence  int     `json:"confidence"`   // 定位置信度(0-100)
	LocType     int     `json:"loc_type"`     // 定位类型: 0-GPS, 1-网络, 2-基站, 3-WiFi
	LocTime     string  `json:"loc_time"`     // 定位时间
	CoordType   string  `json:"coord_type"`   // 坐标类型
	Address     string  `json:"address"`      // 详细地址
	Province    string  `json:"province"`     // 省
	City        string  `json:"city"`         // 市
	District    string  `json:"district"`     // 区
	Street      string  `json:"street"`       // 街道
	StreetNum   string  `json:"street_num"`   // 门牌号
	Country     string  `json:"country"`      // 国家
	CountryCode int     `json:"country_code"` // 国家代码
}

// HardwareLocBody 定位结果body项
type HardwareLocBody struct {
	Type     int    `json:"type"`     // 定位类型: 2-WiFi, 3-基站
	Location string `json:"location"` // 经纬度，格式: "lon,lat"
	Radius   int    `json:"radius"`   // 精度半径(米)
	Country  string `json:"country"`  // 国家
	Province string `json:"province"` // 省
	City     string `json:"city"`     // 市
	CityCode string `json:"citycode"` // 城市编码
	District string `json:"district"` // 区
	Road     string `json:"road"`     // 街道
	Ctime    string `json:"ctime"`    // 时间戳
	Error    int    `json:"error"`    // 错误码
	AdCode   string `json:"adcode"`   // 行政区划代码
}

// HardwareLocResponse 硬件定位响应
type HardwareLocResponse struct {
	ErrCode int                `json:"errcode"` // API错误码，0表示成功
	ErrMsg  string             `json:"msg"`     // API错误信息
	Body    []HardwareLocBody  `json:"body"`    // 定位结果列表
	// 兼容旧字段
	Status    int               `json:"status"`      // 状态码
	Message   string            `json:"message"`     // 状态描述
	Result    HardwareLocResult `json:"result"`      // 定位结果(解析后填充)
	ErrorCode int               `json:"error_code"`  // 错误码
	ErrorMsg  string            `json:"error_msg"`   // 错误信息
}

// WiFiLocRequest WiFi定位请求(简化版)
type WiFiLocRequest struct {
	// WiFi信息列表，格式: ["mac:rssi", "mac:rssi"]
	// 例如: ["54:6c:3e:77:89:ab:-45", "a4:5d:3c:12:34:56:-67"]
	WiFiData []string `json:"wifi_data"`
	// 是否返回详细地址
	NeedAddress bool `json:"need_address,omitempty"`
}

// CellLocRequest 基站定位请求(简化版)
type CellLocRequest struct {
	// 基站信息
	LAC int `json:"lac"` // 位置区域码
	CID int `json:"cid"` // 基站ID
	// 经纬度(可选)
	Lon float64 `json:"lon,omitempty"`
	Lat float64 `json:"lat,omitempty"`
	// 是否返回详细地址
	NeedAddress bool `json:"need_address,omitempty"`
}

// GetLocationByWiFi WiFi定位
// 通过WiFi接入点信息获取位置
func (m *MapsClient) GetLocationByWiFi(req *HardwareLocRequest) (*HardwareLocResponse, error) {
	if len(req.WiFiList) == 0 {
		return nil, fmt.Errorf("WiFi列表不能为空")
	}

	req.LocType = "wifi"
	return m.hardwareLocate(req)
}

// GetLocationByCell 基站定位
// 通过基站信息获取位置
func (m *MapsClient) GetLocationByCell(req *HardwareLocRequest) (*HardwareLocResponse, error) {
	if len(req.CellList) == 0 {
		return nil, fmt.Errorf("基站列表不能为空")
	}

	req.LocType = "cell"
	return m.hardwareLocate(req)
}

// GetLocation 混合定位
// 同时使用WiFi和基站信息进行定位
func (m *MapsClient) GetLocation(req *HardwareLocRequest) (*HardwareLocResponse, error) {
	if len(req.WiFiList) == 0 && len(req.CellList) == 0 {
		return nil, fmt.Errorf("WiFi列表和基站列表不能同时为空")
	}

	if req.LocType == "" {
		req.LocType = "mixed"
	}
	return m.hardwareLocate(req)
}

// hardwareLocate 硬件定位核心方法
func (m *MapsClient) hardwareLocate(req *HardwareLocRequest) (*HardwareLocResponse, error) {
	// 校验必填参数
	if req.IMEI == "" {
		return nil, fmt.Errorf("IMEI不能为空")
	}
	if len(req.CellList) == 0 {
		return nil, fmt.Errorf("基站信息(CellList)不能为空")
	}

	body := m.buildRequestBody(req)
	slog.Debug("[MapsClient] Request", "url", MapsBaseURL+"/locapi/v2", "body", body)

	result, err := m.sendRequest(body)
	if err != nil {
		return nil, err
	}

	// 从body数组提取定位结果
	m.parseResult(result)
	return result, nil
}

// buildRequestBody 构建定位请求体
func (m *MapsClient) buildRequestBody(req *HardwareLocRequest) map[string]any {
	bodyContent := map[string]any{
		"need_rgc": "Y",                                  // 需要逆地理编码
		"imei":     req.IMEI,                             // 设备IMEI
		"cdma":     req.CDMA,                             // 是否CDMA网络
		"ctime":    fmt.Sprintf("%d", time.Now().Unix()), // 当前时间戳
	}

	// 设置接入类型和定位数据
	if len(req.WiFiList) > 0 {
		bodyContent["accesstype"] = 0 // WiFi接入网络
		wifiData := make([]string, 0, len(req.WiFiList))
		for _, wifi := range req.WiFiList {
			wifiData = append(wifiData, fmt.Sprintf("%s,%d", wifi.MAC, wifi.RSSI))
		}
		bodyContent["macs"] = strings.Join(wifiData, ",")
	}

	// 设置基站信息
	if len(req.CellList) > 0 {
		cell := req.CellList[0]
		bodyContent["bts"] = fmt.Sprintf("%d,%d,%d,%d,-63", cell.MCC, cell.MNC, cell.LAC, cell.CID)
	}

	return map[string]any{
		"key":    m.ak,
		"src":    "jueduipai",
		"prod":   "jueduipai",
		"ver":    "1.0",
		"trace":  false,
		"output": "JSON",
		"body":   []map[string]any{bodyContent},
	}
}

// sendRequest 发送定位请求
func (m *MapsClient) sendRequest(body map[string]any) (*HardwareLocResponse, error) {
	var result HardwareLocResponse
	resp, err := m.httpClient.R().
		SetQueryParam("ak", m.ak).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(MapsBaseURL + "/locapi/v2")

	if err != nil {
		slog.Error("[MapsClient] hardwareLocate request failed", "err", err)
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	// 打印原始响应体
	fmt.Printf("[DEBUG] Response Body: %s\n", resp.String())

	if resp.IsError() {
		slog.Error("[MapsClient] hardwareLocate response error", "status", resp.StatusCode(), "body", resp.String())
		return nil, fmt.Errorf("请求错误: %s", resp.String())
	}

	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查API错误
	if result.ErrCode != 0 {
		slog.Error("[MapsClient] hardwareLocate API error", "errcode", result.ErrCode, "msg", result.ErrMsg)
		return &result, fmt.Errorf("定位失败[%d]: %s", result.ErrCode, result.ErrMsg)
	}

	return &result, nil
}

// parseResult 从响应中提取定位结果
func (m *MapsClient) parseResult(result *HardwareLocResponse) {
	if len(result.Body) == 0 {
		return
	}
	locBody := result.Body[0]
	if lat, lon, ok := parseLocation(locBody.Location); ok {
		result.Result.Lon = lon
		result.Result.Lat = lat
	}
	result.Result.Radius = float64(locBody.Radius)
	result.Result.Country = locBody.Country
	result.Result.Province = locBody.Province
	result.Result.City = locBody.City
	result.Result.District = locBody.District
	result.Result.Street = locBody.Road
	result.Result.LocType = locBody.Type
}

// ParseWiFiData 解析WiFi数据字符串
// 输入格式: ["54:6c:3e:77:89:ab,-45", "a4:5d:3c:12:34:56,-67"]
func ParseWiFiData(wifiData []string) []WiFiAccessPoint {
	var list []WiFiAccessPoint
	for i, data := range wifiData {
		// 查找逗号位置
		commaIdx := -1
		for j, c := range data {
			if c == ',' {
				commaIdx = j
				break
			}
		}
		if commaIdx == -1 || commaIdx == 0 || commaIdx == len(data)-1 {
			continue
		}

		mac := data[:commaIdx]
		rssiStr := data[commaIdx+1:]

		var rssi int
		if _, err := fmt.Sscanf(rssiStr, "%d", &rssi); err != nil {
			continue
		}

		list = append(list, WiFiAccessPoint{
			MAC:    mac,
			RSSI:   rssi,
			IsMain: i == 0,
		})
	}
	return list
}

// SimpleWiFiLoc 简化的WiFi定位方法
// wifiData: WiFi数据，格式: ["54:6c:3e:77:89:ab,-45", "a4:5d:3c:12:34:56,-67"]
// imei: 设备IMEI，15位数字
// cell: 基站信息
// cdma: 是否CDMA网络(0或1)
func (m *MapsClient) SimpleWiFiLoc(wifiData []string, imei string, cell BaseStation, cdma int, needAddress bool) (*HardwareLocResponse, error) {
	wifiList := ParseWiFiData(wifiData)
	if len(wifiList) == 0 {
		return nil, fmt.Errorf("无效的WiFi数据")
	}

	return m.GetLocationByWiFi(&HardwareLocRequest{
		WiFiList:    wifiList,
		CellList:    []BaseStation{cell},
		IMEI:        imei,
		CDMA:        cdma,
		NeedAddress: needAddress,
	})
}

// SimpleCellLoc 简化的基站定位方法
// SimpleCellLoc 简化的基站定位方法
// imei: 设备IMEI，15位数字
// cell: 基站信息
// cdma: 是否CDMA网络(0或1)
func (m *MapsClient) SimpleCellLoc(imei string, cell BaseStation, cdma int, needAddress bool) (*HardwareLocResponse, error) {
	return m.GetLocationByCell(&HardwareLocRequest{
		CellList:    []BaseStation{cell},
		IMEI:        imei,
		CDMA:        cdma,
		NeedAddress: needAddress,
	})
}

// DeviceLocRequest 设备定位请求(兼容硬件协议)
type DeviceLocRequest struct {
	Type   string   `json:"type"`             // loc
	Source string   `json:"source"`           // wifi 或 bs
	Data   []string `json:"data,omitempty"`   // WiFi数据
	IMEI   string   `json:"imei"`             // 设备IMEI（必填）
	IMSI   string   `json:"imsi"`             // 设备IMSI（必填，从中解析MCC和MNC）
	CDMA   int      `json:"cdma"`             // 是否CDMA网络（必填）
	LAC    int      `json:"lac"`              // 位置区域码（必填）
	CID    int      `json:"cid"`              // 基站ID（必填）
}

// DeviceLocResponse 设备定位响应
type DeviceLocResponse struct {
	Lat       float64 `json:"lat"`       // 纬度
	Lon       float64 `json:"lon"`       // 经度
	Address   string  `json:"address"`   // 地址
	Province  string  `json:"province"`  // 省
	City      string  `json:"city"`      // 市
	District  string  `json:"district"`  // 区
	Radius    float64 `json:"radius"`    // 精度半径(米)
	LocType   string  `json:"loc_type"`  // 定位类型
	Timestamp int64   `json:"timestamp"` // 定位时间戳
}

// LocateDevice 设备定位(兼容硬件协议格式)
func (m *MapsClient) LocateDevice(req *DeviceLocRequest) (*DeviceLocResponse, error) {
	// 从IMSI解析MCC和MNC
	imsiInfo, err := ParseIMSI(req.IMSI)
	if err != nil {
		return nil, fmt.Errorf("解析IMSI失败: %w", err)
	}

	cell := BaseStation{
		MCC: imsiInfo.MCC,
		MNC: imsiInfo.MNC,
		LAC: req.LAC,
		CID: req.CID,
	}

	var resp *HardwareLocResponse

	switch req.Source {
	case "wifi":
		resp, err = m.SimpleWiFiLoc(req.Data, req.IMEI, cell, req.CDMA, true)
	case "bs":
		resp, err = m.SimpleCellLoc(req.IMEI, cell, req.CDMA, true)
	default:
		return nil, fmt.Errorf("不支持的定位类型: %s", req.Source)
	}

	if err != nil {
		return nil, err
	}

	return &DeviceLocResponse{
		Lat:       resp.Result.Lat,
		Lon:       resp.Result.Lon,
		Address:   resp.Result.Address,
		Province:  resp.Result.Province,
		City:      resp.Result.City,
		District:  resp.Result.District,
		Radius:    resp.Result.Radius,
		LocType:   req.Source,
		Timestamp: time.Now().Unix(),
	}, nil
}

// ConvertToJSON 转换为JSON字符串
func (r *HardwareLocResponse) ConvertToJSON() string {
	data, _ := json.Marshal(r)
	return string(data)
}

// IsSuccess 判断定位是否成功
func (r *HardwareLocResponse) IsSuccess() bool {
	return r.Status == 0
}

// GetAddress 获取格式化地址
func (r *HardwareLocResult) GetAddress() string {
	if r.Address != "" {
		return r.Address
	}

	// 拼接地址
	var parts []string
	if r.Province != "" {
		parts = append(parts, r.Province)
	}
	if r.City != "" && r.City != r.Province {
		parts = append(parts, r.City)
	}
	if r.District != "" {
		parts = append(parts, r.District)
	}
	if r.Street != "" {
		parts = append(parts, r.Street)
	}
	if r.StreetNum != "" {
		parts = append(parts, r.StreetNum)
	}

	result := ""
	for i, part := range parts {
		if i > 0 {
			result += part
		} else {
			result = part
		}
	}
	return result
}

// parseLocation 解析位置字符串 "lon,lat"
func parseLocation(location string) (lat, lon float64, ok bool) {
	parts := strings.Split(location, ",")
	if len(parts) != 2 {
		return 0, 0, false
	}
	if _, err := fmt.Sscanf(parts[0], "%f", &lon); err != nil {
		return 0, 0, false
	}
	if _, err := fmt.Sscanf(parts[1], "%f", &lat); err != nil {
		return 0, 0, false
	}
	return lat, lon, true
}
