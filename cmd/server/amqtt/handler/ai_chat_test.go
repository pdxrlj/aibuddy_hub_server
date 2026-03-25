package handler

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	ai "aibuddy/aiframe/ai_chat"
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"aibuddy/pkg/config"
	"aibuddy/pkg/mqtt"

	paho "github.com/eclipse/paho.mqtt.golang"
)

// mockMessage 实现 paho.Message 接口
type mockMessage struct {
	topic   string
	payload []byte
}

func (m *mockMessage) Topic() string     { return m.topic }
func (m *mockMessage) Payload() []byte   { return m.payload }
func (m *mockMessage) Qos() byte         { return 1 }
func (m *mockMessage) MessageID() uint16 { return 1 }
func (m *mockMessage) Ack()              {}
func (m *mockMessage) Duplicate() bool   { return false }
func (m *mockMessage) Retained() bool    { return false }

// TestAiChatHandler_Chat_End 测试对话结束
// 使用设备号 30:ED:A0:E9:F3:07，开始时间 1774347962
func TestAiChatHandler_Chat_End(t *testing.T) {
	// 初始化配置和数据库
	config.Setup("f:/code/aibuddy_hub_server/config")
	db := model.Conn()
	query.SetDefault(db.GetDB())

	// 用户指定的时间 1774347962
	startTime := int64(1774347962)
	deviceID := "30:ED:A0:E9:F3:07"
	sessionID := "test_session_001"

	// 创建 handler
	handler := NewAiChatHandler()

	// 预设开始时间到缓存（使用用户指定的时间 1774347962）
	// cacheKey 格式: ai_chat:{device_id}:{sid}:{type}
	startKey := fmt.Sprintf("ai_chat:%s:%s:%s", deviceID, sessionID, ai.ChatTypeStart)
	if err := handler.cache.Set(startKey, startTime, time.Hour*24); err != nil {
		t.Fatalf("Set start time to cache failed: %v", err)
	}
	t.Logf("Start time preset to cache: startTime=%d (%s)", startTime, time.Unix(startTime, 0).Format("2006-01-02 15:04:05"))

	// 构造对话结束消息
	endMsg := ai.Chat{
		Type: ai.ChatTypeEnd,
		Sid:  sessionID,
	}

	endPayload, err := json.Marshal(endMsg)
	if err != nil {
		t.Fatalf("Marshal end message failed: %v", err)
	}

	// 创建 mock context
	endCtx := &mqtt.Context{
		Topic:   "aibuddy/" + deviceID + "/ai",
		Payload: endPayload,
		Params:  map[string]string{"device_id": deviceID},
		Message: &mockMessage{topic: "aibuddy/" + deviceID + "/ai", payload: endPayload},
	}

	// 执行测试
	handler.Chat(endCtx)

	t.Logf("Chat end message processed, deviceID=%s, sessionID=%s, startTime=%d", deviceID, sessionID, startTime)

	// 等待异步处理完成
	time.Sleep(5 * time.Second)
}

// 确保 mockMessage 实现了 paho.Message 接口
var _ paho.Message = (*mockMessage)(nil)
