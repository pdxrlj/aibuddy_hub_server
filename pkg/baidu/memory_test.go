package baidu

import (
	"testing"

	"aibuddy/pkg/config"
)

func TestMemory_ClearCharacterPortrait(t *testing.T) {
	config.Setup("../../config")

	memory := NewMemory()

	// 测试清空人物画像
	err := memory.ClearCharacterPortrait(&ClearCharacterPortraitRequest{
		UserID: "30:ED:A0:E9:F2:12",
	})
	if err != nil {
		t.Fatalf("ClearCharacterPortrait failed: %v", err)
	}

	t.Log("ClearCharacterPortrait succeeded")
}

func TestMemory_ClearCharacterPortrait_WithAppID(t *testing.T) {
	config.Setup("../../config")

	memory := NewMemory()

	// 测试指定AppID清空人物画像
	err := memory.ClearCharacterPortrait(&ClearCharacterPortraitRequest{
		AppID:  config.Instance.Baidu.AppID,
		UserID: "test_user_002",
	})
	if err != nil {
		t.Fatalf("ClearCharacterPortrait failed: %v", err)
	}

	t.Log("ClearCharacterPortrait with AppID succeeded")
}

func TestMemory_ClearCharacterPortrait_EmptyUserID(t *testing.T) {
	config.Setup("../../config")

	memory := NewMemory()

	// 测试空UserID（预期会失败）
	err := memory.ClearCharacterPortrait(&ClearCharacterPortraitRequest{
		UserID: "",
	})
	if err != nil {
		t.Logf("Expected error for empty UserID: %v", err)
		return
	}

	t.Log("ClearCharacterPortrait with empty UserID succeeded (unexpected)")
}

func TestMemory_ClearCharacterPortrait_InvalidUserID(t *testing.T) {
	config.Setup("../../config")

	memory := NewMemory()

	// 测试不存在的UserID（API可能返回成功，因为清空空数据也算成功）
	err := memory.ClearCharacterPortrait(&ClearCharacterPortraitRequest{
		UserID: "non_existent_user_" + "test_timestamp",
	})
	if err != nil {
		t.Logf("ClearCharacterPortrait for non-existent user returned error: %v", err)
		return
	}

	t.Log("ClearCharacterPortrait for non-existent user succeeded")
}
