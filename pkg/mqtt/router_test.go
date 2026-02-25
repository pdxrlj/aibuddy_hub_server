package mqtt

import (
	"strings"
	"sync"
	"testing"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
)

func TestExtractParams(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		topic    string
		expected map[string]string
	}{
		{
			name:     "single param",
			pattern:  "device/:id/status",
			topic:    "device/123/status",
			expected: map[string]string{"id": "123"},
		},
		{
			name:     "multiple params",
			pattern:  "device/:id/:type",
			topic:    "device/123/temperature",
			expected: map[string]string{"id": "123", "type": "temperature"},
		},
		{
			name:     "no params",
			pattern:  "device/status",
			topic:    "device/status",
			expected: map[string]string{},
		},
		{
			name:     "param at start",
			pattern:  ":id/device",
			topic:    "123/device",
			expected: map[string]string{"id": "123"},
		},
		{
			name:     "mixed pattern",
			pattern:  "group/:groupId/device/:deviceId",
			topic:    "group/g1/device/d1",
			expected: map[string]string{"groupId": "g1", "deviceId": "d1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractParams(tt.pattern, tt.topic)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d params, got %d", len(tt.expected), len(result))
			}
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("expected %s=%s, got %s=%s", k, v, k, result[k])
				}
			}
		})
	}
}

func TestExtractParamNames(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected []string
	}{
		{
			name:     "single param",
			pattern:  "device/:id/status",
			expected: []string{"id"},
		},
		{
			name:     "multiple params",
			pattern:  "device/:id/:type",
			expected: []string{"id", "type"},
		},
		{
			name:     "no params",
			pattern:  "device/status",
			expected: []string{},
		},
		{
			name:     "param at end",
			pattern:  "device/:id",
			expected: []string{"id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractParamNames(tt.pattern)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d params, got %d", len(tt.expected), len(result))
			}
			for i, v := range tt.expected {
				if i >= len(result) || result[i] != v {
					t.Errorf("expected param[%d]=%s, got %s", i, v, result[i])
				}
			}
		})
	}
}

func TestBuildTopic(t *testing.T) {
	r := &Router{topicPrefix: "myapp"}

	tests := []struct {
		name     string
		prefix   string
		topic    string
		expected string
	}{
		{
			name:     "with prefix",
			prefix:   "myapp",
			topic:    "device/status",
			expected: "myapp/device/status",
		},
		{
			name:     "empty prefix",
			prefix:   "",
			topic:    "device/status",
			expected: "device/status",
		},
		{
			name:     "single level topic",
			prefix:   "myapp",
			topic:    "status",
			expected: "myapp/status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r.topicPrefix = tt.prefix
			result := r.buildTopic(tt.topic)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestConvertToMQTTWildcard(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected string
	}{
		{
			name:     "single param",
			pattern:  "device/:id/status",
			expected: "device/+/status",
		},
		{
			name:     "multiple params",
			pattern:  "device/:id/:type",
			expected: "device/+/+",
		},
		{
			name:     "no params",
			pattern:  "device/status",
			expected: "device/status",
		},
		{
			name:     "param at start",
			pattern:  ":id/device",
			expected: "+/device",
		},
		{
			name:     "mixed pattern",
			pattern:  "group/:groupId/device/:deviceId",
			expected: "group/+/device/+",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToMQTTWildcard(tt.pattern)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestBuildSubscribeTopic(t *testing.T) {
	r := &Router{topicPrefix: "nlaibuddy"}

	tests := []struct {
		name     string
		prefix   string
		pattern  string
		expected string
	}{
		{
			name:     "with prefix and param",
			prefix:   "nlaibuddy",
			pattern:  "device/:id/status",
			expected: "nlaibuddy/device/+/status",
		},
		{
			name:     "without prefix but with param",
			prefix:   "",
			pattern:  "device/:id/status",
			expected: "device/+/status",
		},
		{
			name:     "with prefix no param",
			prefix:   "myapp",
			pattern:  "device/status",
			expected: "myapp/device/status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r.topicPrefix = tt.prefix
			result := r.buildSubscribeTopic(tt.pattern)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestRouterOn(t *testing.T) {
	r := &Router{
		subscribed: make(map[string]bool),
	}

	// 测试注册路由
	routeCount := 0
	handler := func(ctx *Context) {
		routeCount++
	}

	// 模拟注册（不实际订阅，因为没有 mqtt 实例）
	r.mu.Lock()
	route := &Route{Topic: "test/:id", Handler: handler}
	r.routes = append(r.routes, route)
	r.mu.Unlock()

	if len(r.routes) != 1 {
		t.Errorf("expected 1 route, got %d", len(r.routes))
	}
}

func TestContextString(t *testing.T) {
	ctx := &Context{
		Payload: []byte("hello world"),
	}

	if ctx.String() != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", ctx.String())
	}
}

func TestContextBindJSON(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name    string
		payload []byte
		wantErr bool
	}{
		{
			name:    "valid json",
			payload: []byte(`{"name":"test","age":25}`),
			wantErr: false,
		},
		{
			name:    "invalid json",
			payload: []byte(`{"name":invalid}`),
			wantErr: true,
		},
		{
			name:    "empty payload",
			payload: []byte{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &Context{Payload: tt.payload}
			var result TestStruct
			err := ctx.BindJSON(&result)

			if (err != nil) != tt.wantErr {
				t.Errorf("BindJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// --- Mock 实现 ---

// mockClient 模拟 MQTT 客户端
type mockClient struct {
	connected     bool
	subscriptions map[string]paho.MessageHandler
	publishedMsgs []publishedMsg
	mu            sync.Mutex
}

type publishedMsg struct {
	topic   string
	payload any
}

func newMockClient() *mockClient {
	return &mockClient{
		subscriptions: make(map[string]paho.MessageHandler),
	}
}

func (m *mockClient) IsConnected() bool {
	return m.connected
}

func (m *mockClient) IsConnectionOpen() bool {
	return m.connected
}

func (m *mockClient) Connect() paho.Token {
	m.connected = true
	return &mockToken{}
}

func (m *mockClient) Disconnect(quiesce uint) {
	m.connected = false
}

func (m *mockClient) Subscribe(topic string, qos byte, callback paho.MessageHandler) paho.Token {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.subscriptions[topic] = callback
	return &mockToken{}
}

func (m *mockClient) SubscribeMultiple(filters map[string]byte, callback paho.MessageHandler) paho.Token {
	m.mu.Lock()
	defer m.mu.Unlock()
	for topic := range filters {
		m.subscriptions[topic] = callback
	}
	return &mockToken{}
}

func (m *mockClient) Unsubscribe(topics ...string) paho.Token {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range topics {
		delete(m.subscriptions, t)
	}
	return &mockToken{}
}

func (m *mockClient) AddRoute(topic string, callback paho.MessageHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.subscriptions[topic] = callback
}

func (m *mockClient) Publish(topic string, qos byte, retained bool, payload any) paho.Token {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.publishedMsgs = append(m.publishedMsgs, publishedMsg{topic: topic, payload: payload})
	return &mockToken{}
}

// 模拟收到消息
func (m *mockClient) simulateMessage(topic string, payload []byte) {
	m.mu.Lock()
	var handler paho.MessageHandler
	var found bool

	// 先尝试精确匹配
	if h, ok := m.subscriptions[topic]; ok && h != nil {
		handler = h
		found = true
	}

	// 尝试参数模式匹配
	if !found {
		for pattern, h := range m.subscriptions {
			if h == nil {
				continue
			}
			if matchPattern(pattern, topic) {
				handler = h
				found = true
				break
			}
		}
	}
	m.mu.Unlock()

	if found && handler != nil {
		handler(m, &mockMessage{topic: topic, payload: payload})
	}
}

// matchPattern 检查 topic 是否匹配模式（支持 MQTT + 通配符）
func matchPattern(pattern, topic string) bool {
	patternParts := splitPath(pattern)
	topicParts := splitPath(topic)

	if len(patternParts) != len(topicParts) {
		return false
	}

	for i := range len(patternParts) {
		// MQTT + 通配符匹配任意单级
		if patternParts[i] == "+" {
			continue
		}
		// 非通配符部分必须精确匹配
		if patternParts[i] != topicParts[i] {
			return false
		}
	}
	return true
}

func splitPath(path string) []string {
	return strings.Split(path, "/")
}

// OptionsReader 返回空的选项读取器
func (m *mockClient) OptionsReader() paho.ClientOptionsReader {
	return paho.ClientOptionsReader{}
}

// mockToken 实现 paho.Token
type mockToken struct{}

func (t *mockToken) Wait() bool                       { return true }
func (t *mockToken) WaitTimeout(d time.Duration) bool { return true }
func (t *mockToken) Error() error                     { return nil }
func (t *mockToken) Done() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

// mockMessage 实现 paho.Message
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

// --- 订阅发布测试 ---

func TestRouterSubscribeAndMessage(t *testing.T) {
	// 创建 mock 客户端
	mockClt := newMockClient()
	mockClt.Connect()

	// 创建 Mqtt 实例
	m := &Mqtt{Client: mockClt}

	// 创建路由器
	r := NewRouter(m, "test")

	// 记录收到的消息
	var receivedMsg struct {
		topic   string
		payload string
		params  map[string]string
		mu      sync.Mutex
	}

	// 注册路由
	r.On("device/:id/status", func(ctx *Context) {
		receivedMsg.mu.Lock()
		defer receivedMsg.mu.Unlock()
		receivedMsg.topic = ctx.Topic
		receivedMsg.payload = ctx.String()
		receivedMsg.params = ctx.Params
	})

	// 模拟收到消息
	mockClt.simulateMessage("test/device/123/status", []byte(`{"temperature": 25.5}`))

	// 验证
	receivedMsg.mu.Lock()
	defer receivedMsg.mu.Unlock()

	if receivedMsg.topic != "test/device/123/status" {
		t.Errorf("expected topic 'test/device/123/status', got '%s'", receivedMsg.topic)
	}
	if receivedMsg.payload != `{"temperature": 25.5}` {
		t.Errorf("expected payload '{\"temperature\": 25.5}', got '%s'", receivedMsg.payload)
	}
	if receivedMsg.params["id"] != "123" {
		t.Errorf("expected param id='123', got '%s'", receivedMsg.params["id"])
	}
}

func TestRouterPublish(t *testing.T) {
	mockClt := newMockClient()
	mockClt.Connect()

	m := &Mqtt{Client: mockClt}
	r := NewRouter(m, "myapp")

	// 发布消息
	err := r.Publish("device/status", map[string]any{"status": "online"})
	if err != nil {
		t.Errorf("Publish failed: %v", err)
	}

	// 验证发布的消息
	if len(mockClt.publishedMsgs) != 1 {
		t.Errorf("expected 1 published message, got %d", len(mockClt.publishedMsgs))
	}

	msg := mockClt.publishedMsgs[0]
	if msg.topic != "myapp/device/status" {
		t.Errorf("expected topic 'myapp/device/status', got '%s'", msg.topic)
	}
}

func TestContextReply(t *testing.T) {
	mockClt := newMockClient()
	mockClt.Connect()

	m := &Mqtt{Client: mockClt}
	r := NewRouter(m, "")

	// 记录回复
	var replyReceived bool
	r.On("request", func(ctx *Context) {
		err := ctx.Reply("response", "ack")
		if err != nil {
			t.Errorf("Reply failed: %v", err)
		}
		replyReceived = true
	})

	// 模拟收到消息
	mockClt.simulateMessage("request", []byte("test"))

	if !replyReceived {
		t.Error("handler was not called")
	}

	// 验证回复
	if len(mockClt.publishedMsgs) != 1 {
		t.Errorf("expected 1 published message, got %d", len(mockClt.publishedMsgs))
	}
	if mockClt.publishedMsgs[0].topic != "response" {
		t.Errorf("expected reply topic 'response', got '%s'", mockClt.publishedMsgs[0].topic)
	}
}

func TestMiddleware(t *testing.T) {
	mockClt := newMockClient()
	mockClt.Connect()

	m := &Mqtt{Client: mockClt}
	r := NewRouter(m, "")

	var callOrder []string

	// 添加中间件
	r.Use(func(next Handler) Handler {
		return func(ctx *Context) {
			callOrder = append(callOrder, "middleware1-before")
			next(ctx)
			callOrder = append(callOrder, "middleware1-after")
		}
	})

	r.Use(func(next Handler) Handler {
		return func(ctx *Context) {
			callOrder = append(callOrder, "middleware2-before")
			next(ctx)
			callOrder = append(callOrder, "middleware2-after")
		}
	})

	r.On("test", func(ctx *Context) {
		callOrder = append(callOrder, "handler")
	})

	// 模拟收到消息
	mockClt.simulateMessage("test", []byte("data"))

	// 验证中间件调用顺序
	expected := []string{
		"middleware1-before",
		"middleware2-before",
		"handler",
		"middleware2-after",
		"middleware1-after",
	}

	if len(callOrder) != len(expected) {
		t.Errorf("expected %d calls, got %d", len(expected), len(callOrder))
	}

	for i, v := range expected {
		if i >= len(callOrder) || callOrder[i] != v {
			t.Errorf("expected callOrder[%d]='%s', got '%s'", i, v, callOrder[i])
		}
	}
}
