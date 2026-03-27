package baidu

import (
	"testing"
)

func TestParseIMSI(t *testing.T) {
	imsi := "460041888422950"
	info, err := ParseIMSI(imsi)
	if err != nil {
		t.Fatalf("解析IMSI失败: %v", err)
	}

	t.Logf("IMSI: %s", imsi)
	t.Logf("MCC: %d", info.MCC)
	t.Logf("MNC: %d", info.MNC)
	t.Logf("MSIN: %s", info.MSIN)

	if info.MCC != 460 {
		t.Errorf("MCC期望460，实际%d", info.MCC)
	}
	if info.MNC != 4 {
		t.Errorf("MNC期望4，实际%d", info.MNC)
	}
}

func TestWiFiLocation(t *testing.T) {
	// WiFi数据
	wifiData := []string{
		"70:ba:ef:d0:87:91,-42",
		"70:ba:ef:d1:0e:01,-45",
	}

	// 必填参数
	imei := "861540085739384" // 设备IMEI
	imsi := "460041888422950" // 设备IMSI
	cdma := 0

	// 从IMSI解析MCC和MNC
	imsiInfo, err := ParseIMSI(imsi)
	if err != nil {
		t.Fatalf("解析IMSI失败: %v", err)
	}

	cell := BaseStation{
		MCC: imsiInfo.MCC,
		MNC: imsiInfo.MNC,
		LAC: 4189,
		CID: 8869,
	}

	// 创建客户端
	client := NewMapsClientWithAK("hk04Hv2nIJQb7dJo9U0KulylI3eUlfBL", "122590536")

	// 执行WiFi定位
	resp, err := client.SimpleWiFiLoc(wifiData, imei, cell, cdma, true)
	if err != nil {
		t.Fatalf("WiFi定位失败: %v", err)
	}

	// 打印原始响应
	t.Logf("原始响应: %s", resp.ConvertToJSON())

	// 检查响应
	if !resp.IsSuccess() {
		t.Fatalf("定位失败: status=%d, message=%s", resp.Status, resp.Message)
	}

	// 打印定位结果
	t.Logf("定位成功!")
	t.Logf("纬度: %f", resp.Result.Lat)
	t.Logf("经度: %f", resp.Result.Lon)
	t.Logf("精度半径: %.2f 米", resp.Result.Radius)
	t.Logf("置信度: %d", resp.Result.Confidence)
	t.Logf("地址: %s", resp.Result.GetAddress())
	t.Logf("省: %s", resp.Result.Province)
	t.Logf("市: %s", resp.Result.City)
	t.Logf("区: %s", resp.Result.District)
}
