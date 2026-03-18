package baidu

import (
	"testing"

	"aibuddy/pkg/config"
)

func TestSwitchRole_SwitchSceneRole(t *testing.T) {
	config.Setup("../../config")

	sr := NewSwitchRole()

	// 测试切换角色（需要有效的实例ID）
	err := sr.SwitchSceneRole(&SwitchRoleRequest{
		AiAgentInstanceID: 2756212200374272, // 需要替换为有效的实例ID
		SceneRole:         "奶龙李白形态",
	})
	if err != nil {
		t.Logf("SwitchSceneRole failed (expected if instance not exists): %v", err)
		return
	}

	t.Log("SwitchSceneRole succeeded")
}

func TestSwitchRole_SwitchSceneRole_WithTTS(t *testing.T) {
	config.Setup("../../config")

	sr := NewSwitchRole()

	// 测试切换角色并指定TTS配置
	err := sr.SwitchSceneRole(&SwitchRoleRequest{
		AiAgentInstanceID: 2756671157895168, // 需要替换为有效的实例ID
		SceneRole:         "奶龙斯坦",
		TTS:               `DEFAULT{"vcn":"1000454"}`,
		TTSSayHi:          "你好，我是一名英语口语老师",
	})
	if err != nil {
		t.Logf("SwitchSceneRole with TTS failed (expected if instance not exists): %v", err)
		return
	}

	t.Log("SwitchSceneRole with TTS succeeded")
}

func TestSwitchRole_SwitchSceneRole_EmptySceneRole(t *testing.T) {
	config.Setup("../../config")

	sr := NewSwitchRole()

	// 测试空角色名（预期失败）
	err := sr.SwitchSceneRole(&SwitchRoleRequest{
		AiAgentInstanceID: 123456,
		SceneRole:         "",
	})
	if err != nil {
		t.Logf("Expected error for empty scene role: %v", err)
		return
	}

	t.Log("SwitchSceneRole with empty scene role succeeded (unexpected)")
}
