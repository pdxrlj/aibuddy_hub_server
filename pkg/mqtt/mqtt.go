// Package mqtt 提供MQTT客户端连接和管理功能
package mqtt

import (
	"aibuddy/pkg/helpers"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const defaultReconnectDelay = 5 * time.Second

// Instance 全局MQTT实例
var Instance *Mqtt

// Mqtt MQTT实例
type Mqtt struct {
	Client mqtt.Client
}

// IsConnected 检查是否已连接
func (m *Mqtt) IsConnected() bool {
	return m != nil && m.Client != nil && m.Client.IsConnected()
}

// Disconnect 断开连接
func (m *Mqtt) Disconnect() {
	if m != nil && m.Client != nil && m.Client.IsConnected() {
		m.Client.Disconnect(250) // 250ms 等待时间
		slog.Info("Disconnected from MQTT broker")
	}
}

// Publish 发布消息
func (m *Mqtt) Publish(topic string, qos byte, retained bool, payload any) error {
	if !m.IsConnected() {
		return fmt.Errorf("mqtt client not connected")
	}
	token := m.Client.Publish(topic, qos, retained, payload)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// Subscribe 订阅主题
func (m *Mqtt) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) error {
	if !m.IsConnected() {
		return fmt.Errorf("mqtt client not connected")
	}
	token := m.Client.Subscribe(topic, qos, callback)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// Unsubscribe 取消订阅
func (m *Mqtt) Unsubscribe(topics ...string) error {
	if !m.IsConnected() {
		return fmt.Errorf("mqtt client not connected")
	}
	token := m.Client.Unsubscribe(topics...)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// Config MQTT配置
type Config struct {
	Aliyun         *AliyunConfig    `json:"aliyun" mapstructure:"aliyun"`
	URL            string           `json:"url" mapstructure:"url"`
	ClientIDPrefix string           `json:"client_id_prefix" mapstructure:"client_id_prefix"`
	InstanceID     string           `json:"instance_id" mapstructure:"instance_id"`
	TopicPrefix    string           `json:"topic_prefix" mapstructure:"topic_prefix"`
	CleanSession   bool             `json:"clean_session" mapstructure:"clean_session"`
	KeepAlive      time.Duration    `json:"keep_alive" mapstructure:"keep_alive"`
	Reconnect      *ReconnectConfig `json:"reconnect" mapstructure:"reconnect"`
}

// AliyunConfig 阿里云配置
type AliyunConfig struct {
	AccessKeyID     string `json:"access_key_id" mapstructure:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret" mapstructure:"access_key_secret"`
}

// ReconnectConfig 重连配置
type ReconnectConfig struct {
	Delay time.Duration `json:"delay" mapstructure:"delay"`
}

// Connect 创建MQTT实例
func Connect(cfg *Config) (*Mqtt, error) {
	if cfg == nil {
		return nil, fmt.Errorf("mqtt config is nil")
	}

	clientID := generateClientID(cfg.ClientIDPrefix)
	username, password, err := GenerateAKSignature(cfg, clientID)
	if err != nil {
		return nil, err
	}

	opts := buildClientOptions(cfg, clientID, username, password)
	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	Instance = &Mqtt{Client: client}
	return Instance, nil
}

func generateClientID(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, helpers.GenerateUUID(10))
}

func buildClientOptions(cfg *Config, clientID, username, password string) *mqtt.ClientOptions {
	reconnectDelay := defaultReconnectDelay
	if cfg.Reconnect != nil && cfg.Reconnect.Delay > 0 {
		reconnectDelay = cfg.Reconnect.Delay
	}

	opts := mqtt.NewClientOptions().
		AddBroker(cfg.URL).
		SetUsername(username).
		SetPassword(password).
		SetClientID(clientID).
		SetCleanSession(cfg.CleanSession).
		SetKeepAlive(cfg.KeepAlive).
		SetConnectRetry(true).
		SetAutoReconnect(true).
		SetAutoAckDisabled(true).
		SetConnectRetryInterval(reconnectDelay).
		SetConnectionAttemptHandler(func(broker *url.URL, tlsCfg *tls.Config) *tls.Config {
			slog.Info("Connecting to MQTT broker", "broker", broker.String())
			return tlsCfg
		}).
		SetReconnectingHandler(func(_ mqtt.Client, _ *mqtt.ClientOptions) {
			slog.Info("Reconnecting to MQTT broker")
		}).
		SetOnConnectHandler(func(_ mqtt.Client) {
			slog.Info("Connected to MQTT broker")
		}).
		SetConnectionLostHandler(func(_ mqtt.Client, err error) {
			slog.Error("Connection lost to MQTT broker", "error", err)
		})

	return opts
}

// GetTopic 获取主题
func GetTopic(prefix, topic string) string {
	if prefix == "" {
		return strings.TrimPrefix(topic, "/")
	}
	return strings.TrimPrefix(fmt.Sprintf("%s/%s", prefix, topic), "/")
}

// GenerateAKSignature 生成AK签名
func GenerateAKSignature(cfg *Config, clientID string) (string, string, error) {
	if cfg == nil {
		return "", "", fmt.Errorf("mqtt config is nil")
	}
	if cfg.Aliyun == nil {
		return "", "", fmt.Errorf("aliyun config is nil")
	}
	if cfg.InstanceID == "" {
		return "", "", fmt.Errorf("instance id is empty")
	}

	username, password, err := GenerateAliyunMQTTAuth(
		clientID,
		cfg.Aliyun.AccessKeyID,
		cfg.Aliyun.AccessKeySecret,
		cfg.InstanceID,
	)
	if err != nil {
		return "", "", err
	}
	return username, password, nil
}
