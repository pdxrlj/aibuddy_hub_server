// Package mqtt 提供MQTT客户端路由功能
package mqtt

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	paho "github.com/eclipse/paho.mqtt.golang"
)

// Handler 消息处理函数
type Handler func(ctx *Context)

// Middleware 中间件函数
type Middleware func(Handler) Handler

// Context MQTT消息上下文
type Context struct {
	Topic   string
	Payload []byte
	Client  paho.Client
	Message paho.Message
	Params  map[string]string // 路由参数
	router  *Router           // 路由器引用
}

// Reply 回复到指定主题 不带前缀，
func (c *Context) Reply(topic string, payload any) error {
	token := c.Client.Publish(topic, 1, false, payload)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// ReplyWithPrefix 回复到指定主题 自动添加前缀
func (c *Context) ReplyWithPrefix(topic string, payload any) error {
	fullTopic := c.router.buildTopic(topic)
	return c.Reply(fullTopic, payload)
}

// JSON 解析 JSON payload
func (c *Context) JSON(v any) error {
	return json.Unmarshal(c.Payload, v)
}

// String 获取字符串 payload
func (c *Context) String() string {
	return string(c.Payload)
}

// BindJSON 绑定 JSON 并返回错误信息
func (c *Context) BindJSON(v any) error {
	if err := json.Unmarshal(c.Payload, v); err != nil {
		return fmt.Errorf("invalid JSON payload: %w", err)
	}
	return nil
}

// Route 路由定义
type Route struct {
	Topic   string
	Handler Handler
}

// Router MQTT路由器
type Router struct {
	mu          sync.RWMutex
	routes      []*Route
	middlewares []Middleware
	mqtt        *Mqtt
	topicPrefix string
	subscribed  map[string]bool // 已订阅的 topic
	debug       bool            // 调试模式
}

// RouterOption 路由器配置选项
type RouterOption func(*Router)

// WithDebug 启用调试模式
func WithDebug(debug bool) RouterOption {
	return func(r *Router) {
		r.debug = debug
	}
}

// NewRouter 创建路由器
func NewRouter(mqttInstance *Mqtt, topicPrefix string, opts ...RouterOption) *Router {
	r := &Router{
		mqtt:        mqttInstance,
		topicPrefix: topicPrefix,
		subscribed:  make(map[string]bool),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Use 添加中间件 必须在 On 之前调用
func (r *Router) Use(middleware Middleware) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.middlewares = append(r.middlewares, middleware)
}

// Handle 注册路由并自动订阅
func (r *Router) Handle(topic string, handler Handler) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	route := &Route{Topic: topic, Handler: handler}
	r.routes = append(r.routes, route)

	// 自动订阅（将 :param 转换为 MQTT 的 + 通配符）
	subscribeTopic := r.buildSubscribeTopic(topic)
	if !r.subscribed[subscribeTopic] {
		if err := r.subscribeTopic(subscribeTopic, route); err != nil {
			return fmt.Errorf("subscribe %s failed: %w", subscribeTopic, err)
		}
		r.subscribed[subscribeTopic] = true
		slog.Info("[MQTT] Subscribed", "topic", subscribeTopic, "pattern", topic)
	}
	return nil
}

// On 注册路由
func (r *Router) On(topic string, handler Handler) {
	if err := r.Handle(topic, handler); err != nil {
		slog.Error("[MQTT] Failed to register route", "topic", topic, "error", err)
	}
}

// PrintRoutes 打印所有路由信息
func (r *Router) PrintRoutes() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	prefix := r.topicPrefix
	if prefix == "" {
		prefix = "(none)"
	}

	slog.Info("========== MQTT Routes ==========")
	slog.Info("Config", "prefix", prefix, "middlewares", len(r.middlewares), "routes", len(r.routes), "debug", r.debug)

	for i, route := range r.routes {
		subscribeTopic := r.buildSubscribeTopic(route.Topic)
		params := extractParamNames(route.Topic)
		slog.Info("Route", "index", i+1, "pattern", route.Topic, "subscribe_topic", subscribeTopic, "params", params)
	}
	slog.Info("=================================")
}

// extractParamNames 提取路由参数名
func extractParamNames(pattern string) []string {
	var params []string
	for part := range strings.SplitSeq(pattern, "/") {
		if name, ok := strings.CutPrefix(part, ":"); ok {
			params = append(params, name)
		}
	}
	return params
}

func (r *Router) subscribeTopic(topic string, route *Route) error {
	fullPattern := r.buildTopic(route.Topic)

	return r.mqtt.Subscribe(topic, 1, func(client paho.Client, msg paho.Message) {
		ctx := &Context{
			Topic:   msg.Topic(),
			Payload: msg.Payload(),
			Client:  client,
			Message: msg,
			Params:  extractParams(fullPattern, msg.Topic()),
			router:  r,
		}

		// 应用中间件
		handler := route.Handler
		for i := len(r.middlewares) - 1; i >= 0; i-- {
			handler = r.middlewares[i](handler)
		}

		handler(ctx)
	})
}

func (r *Router) buildTopic(topic string) string {
	if r.topicPrefix == "" {
		return topic
	}
	return r.topicPrefix + "/" + topic
}

// buildSubscribeTopic 构建订阅用的 topic（将 :param 转换为 MQTT 的 + 通配符）
func (r *Router) buildSubscribeTopic(topic string) string {
	// 将 :param 转换为 +
	mqttTopic := convertToMQTTWildcard(topic)
	return r.buildTopic(mqttTopic)
}

// convertToMQTTWildcard 将路由参数 :param 转换为 MQTT 通配符 +
func convertToMQTTWildcard(pattern string) string {
	parts := strings.Split(pattern, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			parts[i] = "+"
		}
	}
	return strings.Join(parts, "/")
}

// Publish 发布消息 自动添加前缀
func (r *Router) Publish(topic string, payload any) error {
	return r.mqtt.Publish(r.buildTopic(topic), 1, false, payload)
}

// extractParams 从 topic 中提取参数
func extractParams(pattern, topic string) map[string]string {
	params := make(map[string]string)
	patternParts := strings.Split(pattern, "/")
	topicParts := strings.Split(topic, "/")

	for i := 0; i < len(patternParts) && i < len(topicParts); i++ {
		if paramName, ok := strings.CutPrefix(patternParts[i], ":"); ok {
			params[paramName] = topicParts[i]
		}
	}
	return params
}

var defaultRouter *Router

// InitRouter 初始化默认路由器
func InitRouter(mqttInstance *Mqtt, topicPrefix string, opts ...RouterOption) *Router {
	defaultRouter = NewRouter(mqttInstance, topicPrefix, opts...)
	return defaultRouter
}

// GetRouter 获取默认路由器
func GetRouter() *Router {
	if defaultRouter == nil {
		panic("mqtt router not initialized, call InitRouter first")
	}
	return defaultRouter
}

// PrintRoutes 打印默认路由器的路由信息
func PrintRoutes() {
	GetRouter().PrintRoutes()
}
