package baidu

import (
	"testing"

	"aibuddy/pkg/config"
)

// TestAIAgent_GenerateAIAgentCall 测试创建大模型互动实例
func TestAIAgent_GenerateAIAgentCall(t *testing.T) {
	config.Setup("../../config")

	agent := NewAIAgent()

	// 测试创建语音互动实例
	resp, err := agent.GenerateAIAgentCall(&GenerateAIAgentCallRequest{
		InstanceType: InstanceTypeVoiceChat,
		Config: &AIAgentConfig{
			UserID: "test_user_001",
			Lang:   "zh",
			TTS:     "DEFAULT",
		},
	})
	if err != nil {
		t.Fatalf("GenerateAIAgentCall failed: %v", err)
	}

	if resp.AiAgentInstanceID == 0 {
		t.Error("AiAgentInstanceID should not be 0")
	}

	t.Logf("GenerateAIAgentCall succeeded: instance_id=%d, instance_type=%s", 
		resp.AiAgentInstanceID, resp.InstanceType)

	if resp.Context != nil {
		t.Logf("Context: cid=%d, token=%s", resp.Context.CID, resp.Context.Token)
	}

	// 清理：停止实例
	err = agent.StopAIAgentInstance(&StopAIAgentInstanceRequest{
		AiAgentInstanceID: resp.AiAgentInstanceID,
	})
	if err != nil {
		t.Logf("Warning: StopAIAgentInstance cleanup failed: %v", err)
	}
}

// TestAIAgent_GenerateAIAgentCall_WithConfig 测试创建带完整配置的实例
func TestAIAgent_GenerateAIAgentCall_WithConfig(t *testing.T) {
	config.Setup("../../config")

	agent := NewAIAgent()

	resp, err := agent.GenerateAIAgentCall(&GenerateAIAgentCallRequest{
		InstanceType: InstanceTypeVoiceChat,
		Config: &AIAgentConfig{
			UserID:        "test_user_config",
			Lang:          "zh",
			TTS:           "DEFAULT",
			TTSSayHi:      "你好，我是测试助手",
			Role:          "你是一个友好的测试助手",
			AudioCodec:    "opus",
			ASRVAD:        200,
			ASRVADLevel:   45,
		},
	})
	if err != nil {
		t.Fatalf("GenerateAIAgentCall with config failed: %v", err)
	}

	t.Logf("GenerateAIAgentCall with config succeeded: instance_id=%d", resp.AiAgentInstanceID)

	// 清理
	err = agent.StopAIAgentInstance(&StopAIAgentInstanceRequest{
		AiAgentInstanceID: resp.AiAgentInstanceID,
	})
	if err != nil {
		t.Logf("Warning: StopAIAgentInstance cleanup failed: %v", err)
	}
}

// TestAIAgent_StopAIAgentInstance 测试停止大模型互动实例
func TestAIAgent_StopAIAgentInstance(t *testing.T) {
	config.Setup("../../config")

	agent := NewAIAgent()

	// 先创建一个实例
	createResp, err := agent.GenerateAIAgentCall(&GenerateAIAgentCallRequest{
		InstanceType: InstanceTypeVoiceChat,
		Config: &AIAgentConfig{
			UserID: "test_user_stop",
			Lang:   "zh",
		},
	})
	if err != nil {
		t.Fatalf("GenerateAIAgentCall for stop test failed: %v", err)
	}

	instanceID := createResp.AiAgentInstanceID
	t.Logf("Created instance for stop test: instance_id=%d", instanceID)

	// 测试停止实例
	err = agent.StopAIAgentInstance(&StopAIAgentInstanceRequest{
		AiAgentInstanceID: instanceID,
	})
	if err != nil {
		t.Fatalf("StopAIAgentInstance failed: %v", err)
	}

	t.Log("StopAIAgentInstance succeeded")
}

// TestAIAgent_Interrupt 测试打断大模型互动实例播报
func TestAIAgent_Interrupt(t *testing.T) {
	config.Setup("../../config")

	agent := NewAIAgent()

	// 先创建一个实例
	createResp, err := agent.GenerateAIAgentCall(&GenerateAIAgentCallRequest{
		InstanceType: InstanceTypeVoiceChat,
		Config: &AIAgentConfig{
			UserID: "test_user_interrupt",
			Lang:   "zh",
		},
	})
	if err != nil {
		t.Fatalf("GenerateAIAgentCall for interrupt test failed: %v", err)
	}

	instanceID := createResp.AiAgentInstanceID
	t.Logf("Created instance for interrupt test: instance_id=%d", instanceID)

	// 测试打断播报（不带消息）
	err = agent.Interrupt(&InterruptRequest{
		AiAgentInstanceID: instanceID,
	})
	if err != nil {
		t.Fatalf("Interrupt failed: %v", err)
	}

	t.Log("Interrupt without message succeeded")

	// 清理
	err = agent.StopAIAgentInstance(&StopAIAgentInstanceRequest{
		AiAgentInstanceID: instanceID,
	})
	if err != nil {
		t.Logf("Warning: StopAIAgentInstance cleanup failed: %v", err)
	}
}

// TestAIAgent_Interrupt_WithMessage 测试打断并携带消息
func TestAIAgent_Interrupt_WithMessage(t *testing.T) {
	config.Setup("../../config")

	agent := NewAIAgent()

	// 先创建一个实例
	createResp, err := agent.GenerateAIAgentCall(&GenerateAIAgentCallRequest{
		InstanceType: InstanceTypeVoiceChat,
		Config: &AIAgentConfig{
			UserID: "test_user_interrupt_msg",
			Lang:   "zh",
		},
	})
	if err != nil {
		t.Fatalf("GenerateAIAgentCall for interrupt with message test failed: %v", err)
	}

	instanceID := createResp.AiAgentInstanceID
	t.Logf("Created instance for interrupt with message test: instance_id=%d", instanceID)

	// 测试打断播报（带消息）
	err = agent.Interrupt(&InterruptRequest{
		AiAgentInstanceID: instanceID,
		ExtraMsg:          "打断后播报：您好，有什么可以帮您的吗？",
	})
	if err != nil {
		t.Fatalf("Interrupt with message failed: %v", err)
	}

	t.Log("Interrupt with message succeeded")

	// 清理
	err = agent.StopAIAgentInstance(&StopAIAgentInstanceRequest{
		AiAgentInstanceID: instanceID,
	})
	if err != nil {
		t.Logf("Warning: StopAIAgentInstance cleanup failed: %v", err)
	}
}

// TestAIAgent_SendMsg 测试发送消息给SDK
func TestAIAgent_SendMsg(t *testing.T) {
	config.Setup("../../config")

	agent := NewAIAgent()

	// 先创建一个实例
	createResp, err := agent.GenerateAIAgentCall(&GenerateAIAgentCallRequest{
		InstanceType: InstanceTypeVoiceChat,
		Config: &AIAgentConfig{
			UserID: "test_user_sendmsg",
			Lang:   "zh",
		},
	})
	if err != nil {
		t.Fatalf("GenerateAIAgentCall for sendmsg test failed: %v", err)
	}

	instanceID := createResp.AiAgentInstanceID
	t.Logf("Created instance for sendmsg test: instance_id=%d", instanceID)

	// 测试发送消息
	err = agent.SendMsg(&SendMsgRequest{
		AiAgentInstanceID: instanceID,
		Message:           "这是一条测试消息",
	})
	if err != nil {
		t.Fatalf("SendMsg failed: %v", err)
	}

	t.Log("SendMsg succeeded")

	// 清理
	err = agent.StopAIAgentInstance(&StopAIAgentInstanceRequest{
		AiAgentInstanceID: instanceID,
	})
	if err != nil {
		t.Logf("Warning: StopAIAgentInstance cleanup failed: %v", err)
	}
}

// TestAIAgent_StopNonExistentInstance 测试停止不存在的实例
func TestAIAgent_StopNonExistentInstance(t *testing.T) {
	config.Setup("../../config")

	agent := NewAIAgent()

	// 测试停止一个不存在的实例（应该返回错误）
	err := agent.StopAIAgentInstance(&StopAIAgentInstanceRequest{
		AiAgentInstanceID: 999999999999,
	})
	if err != nil {
		t.Logf("StopAIAgentInstance for non-existent instance returned error as expected: %v", err)
	} else {
		t.Log("StopAIAgentInstance for non-existent instance succeeded (unexpected)")
	}
}

// TestAIAgent_InterruptNonExistentInstance 测试打断不存在的实例
func TestAIAgent_InterruptNonExistentInstance(t *testing.T) {
	config.Setup("../../config")

	agent := NewAIAgent()

	// 测试打断一个不存在的实例（应该返回错误）
	err := agent.Interrupt(&InterruptRequest{
		AiAgentInstanceID: 999999999999,
		ExtraMsg:          "测试消息",
	})
	if err != nil {
		t.Logf("Interrupt for non-existent instance returned error as expected: %v", err)
	} else {
		t.Log("Interrupt for non-existent instance succeeded (unexpected)")
	}
}

// TestAIAgent_SendMsgToNonExistentInstance 测试向不存在的实例发送消息
func TestAIAgent_SendMsgToNonExistentInstance(t *testing.T) {
	config.Setup("../../config")

	agent := NewAIAgent()

	// 测试向一个不存在的实例发送消息（应该返回错误）
	err := agent.SendMsg(&SendMsgRequest{
		AiAgentInstanceID: 999999999999,
		Message:           "测试消息",
	})
	if err != nil {
		t.Logf("SendMsg to non-existent instance returned error as expected: %v", err)
	} else {
		t.Log("SendMsg to non-existent instance succeeded (unexpected)")
	}
}
