package quec

import (
	"testing"
	"time"
)

const (
	testAppKey    = "a835Fs8ON1BjE4MV"
	testSecretKey = "t2SiD64l"
	testIccid     = "898607B8102580151949"
)

func TestGetMonthFlow(t *testing.T) {
	month := time.Now().Format("200601")

	result, err := GetMonthFlow(testAppKey, testSecretKey, month, testIccid)
	if err != nil {
		t.Fatalf("GetMonthFlow failed: %v", err)
	}

	t.Logf("ResultCode: %d, ErrorMessage: %s", result.ResultCode, result.ErrorMessage)

	if result.ResultCode != 0 {
		t.Logf("Response code: %d, ErrorMessage: %s", result.ResultCode, result.ErrorMessage)
		return
	}

	t.Logf("MonthFlow: %.3f MB", result.MonthFlow)
	t.Logf("CycleFlow: %.3f MB", result.CycleFlow)
}

func TestGetBatchMonthFlow(t *testing.T) {
	month := time.Now().Format("200601")

	result, err := GetBatchMonthFlow(testAppKey, testSecretKey, month, testIccid)
	if err != nil {
		t.Fatalf("GetBatchMonthFlow failed: %v", err)
	}

	t.Logf("ResultCode: %d, ErrorMessage: %s", result.ResultCode, result.ErrorMessage)

	if result.ResultCode != 0 {
		t.Logf("Response code: %d, ErrorMessage: %s", result.ResultCode, result.ErrorMessage)
		return
	}

	for i, item := range result.Data {
		t.Logf("[%d] Number: %s, MonthFlow: %.3f MB, CycleFlow: %.3f MB",
			i, item.Number, item.MonthFlow, item.CycleFlow)
	}
}

func TestGetMonthFlowList(t *testing.T) {
	month := time.Now().Format("200601")

	result, err := GetMonthFlowList(testAppKey, testSecretKey, month, 1, 10)
	if err != nil {
		t.Fatalf("GetMonthFlowList failed: %v", err)
	}

	t.Logf("ResultCode: %d, ErrorMessage: %s", result.ResultCode, result.ErrorMessage)

	if result.ResultCode != 0 {
		t.Logf("Response code: %d, ErrorMessage: %s", result.ResultCode, result.ErrorMessage)
		return
	}

	t.Logf("Page: %d/%d, PageSize: %d, TotalCount: %d",
		result.Page.Page, result.Page.TotalPage, result.Page.PageSize, result.Page.TotalCount)

	for i, item := range result.Flows {
		t.Logf("[%d] MSISDN: %s, ICCID: %s, MonthFlow: %.3f MB, Flow: %.3f MB",
			i, item.Msisdn, item.Iccid, item.MonthFlow, item.Flow)
	}
}

func TestGetDayFlowList(t *testing.T) {
	month := time.Now().Format("200601")

	result, err := GetDayFlowList(testAppKey, testSecretKey, month, testIccid)
	if err != nil {
		t.Fatalf("GetDayFlowList failed: %v", err)
	}

	t.Logf("ResultCode: %d, ErrorMessage: %s", result.ResultCode, result.ErrorMessage)

	if result.ResultCode != 0 {
		t.Logf("Response code: %d, ErrorMessage: %s", result.ResultCode, result.ErrorMessage)
		return
	}

	for i, item := range result.Flows {
		t.Logf("[%d] Date: %s, Flow: %.3f MB", i, item.Date, item.Flow)
	}
}
