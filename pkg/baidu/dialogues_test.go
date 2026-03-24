package baidu

import (
	"testing"
	"time"

	"aibuddy/pkg/config"
)

func TestGetDialogues(t *testing.T) {
	config.Setup("../../config")

	dialogues := NewDialogues()

	// 查询最近7天的对话记录
	now := time.Now()
	beginTime := now.Add(-7 * 24 * time.Hour).Unix()
	endTime := now.Unix()

	resp, err := dialogues.GetDialogues(&DialoguesRequest{
		UserID:    "30:ED:A0:E9:F3:07",
		PageNo:    1,
		PageSize:  1000,
		BeginTime: beginTime,
		EndTime:   endTime,
	})
	if err != nil {
		t.Fatalf("GetDialogues failed: %v", err)
	}

	t.Logf("PageNo: %d, PageSize: %d", resp.PageNo, resp.PageSize)
	for i, d := range resp.Data {
		t.Logf("Dialogue %d: Type=%s, Timestamp=%d, Text=%s",
			i+1, d.Type, d.Timestamp, d.Text)
	}
}

func TestGetDialogues_Empty(t *testing.T) {
	config.Setup("../../config")

	dialogues := NewDialogues()

	// 查询一个不存在用户的对话记录
	now := time.Now()
	beginTime := now.Add(-1 * 24 * time.Hour).Unix()
	endTime := now.Unix()

	resp, err := dialogues.GetDialogues(&DialoguesRequest{
		UserID:    "non_existent_user_" + time.Now().Format("20060102150405"),
		PageNo:    1,
		PageSize:  10,
		BeginTime: beginTime,
		EndTime:   endTime,
	})
	if err != nil {
		t.Fatalf("GetDialogues failed: %v", err)
	}

	t.Logf("PageNo: %d, PageSize: %d, DataCount: %d", resp.PageNo, resp.PageSize, len(resp.Data))
}
