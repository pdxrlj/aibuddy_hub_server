// Package config 提供配置管理功能
package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Instance 全局配置实例
var Instance *Config

// Config 应用程序配置
type Config struct {
	App     *AppConfig     `json:"app" mapstructure:"app"`
	Storage *StorageConfig `json:"storage" mapstructure:"storage"`
	Agent   *AgentConfig   `json:"agent" mapstructure:"agent"`
	Tracer  *TracerConfig  `json:"tracer" mapstructure:"tracer"`
	Aliyun  *AliyunConfig  `json:"aliyun" mapstructure:"aliyun"`
	Wechat  *WechatConfig  `json:"wechat" mapstructure:"wechat"`
}

// AppConfig 应用配置
type AppConfig struct {
	Name      string `json:"name" mapstructure:"name"`
	Host      string `json:"host" mapstructure:"host"`
	Port      string `json:"port" mapstructure:"port"`
	LogLevel  string `json:"log_level" mapstructure:"log_level"`
	AppSecret string `json:"app_secret" mapstructure:"app_secret"`
}

// WechatConfig 微信配置
type WechatConfig struct {
	AppID     string `json:"app_id" mapstructure:"app_id"`
	AppSecret string `json:"app_secret" mapstructure:"app_secret"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Database *DatabaseConfig `json:"database" mapstructure:"database"`
	Redis    *RedisConfig    `json:"redis" mapstructure:"redis"`
	Flash    *FlashConfig    `json:"flash" mapstructure:"flash"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Name     string `json:"name" mapstructure:"name"`
	User     string `json:"user" mapstructure:"user"`
	Password string `json:"password" mapstructure:"password"`
	Host     string `json:"host" mapstructure:"host"`
	Port     int    `json:"port" mapstructure:"port"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Username string `json:"username" mapstructure:"username"`
	Password string `json:"password" mapstructure:"password"`
	Host     string `json:"host" mapstructure:"host"`
	Port     int    `json:"port" mapstructure:"port"`
	DB       int    `json:"db" mapstructure:"db"`
}

// FlashConfig 闪存配置
type FlashConfig struct {
	Use string `json:"use" mapstructure:"use"`
}

// AgentConfig Agent配置
type AgentConfig struct {
	Model *AgentModelConfig `json:"model" mapstructure:"model"`
}

// AgentModelConfig Agent模型配置
type AgentModelConfig struct {
	ChatModel *ChatModelConfig `json:"chat_model" mapstructure:"chat_model"`
	WorkModel *WorkModelConfig `json:"work_model" mapstructure:"work_model"`
}

// ChatModelConfig 聊天模型配置
type ChatModelConfig struct {
	ModelName string `json:"model_name" mapstructure:"model_name"`
	APIKey    string `json:"api_key" mapstructure:"api_key"`
	APIURL    string `json:"api_url" mapstructure:"api_url"`
}

// WorkModelConfig 工作模型配置
type WorkModelConfig struct {
	ModelName string `json:"model_name" mapstructure:"model_name"`
	APIKey    string `json:"api_key" mapstructure:"api_key"`
	APIURL    string `json:"api_url" mapstructure:"api_url"`
}

// TracerConfig 追踪配置
type TracerConfig struct {
	ServiceName string `json:"service_name" mapstructure:"service_name"`
	Endpoint    string `json:"endpoint" mapstructure:"endpoint"`
}

// AliyunConfig Aliyun配置
type AliyunConfig struct {
	AccessKeyID     string      `json:"access_key_id" mapstructure:"access_key_id"`
	AccessKeySecret string      `json:"access_key_secret" mapstructure:"access_key_secret"`
	Mqtt            *MqttConfig `json:"mqtt" mapstructure:"mqtt"`
}

// MqttConfig Mqtt配置
type MqttConfig struct {
	URL            string           `json:"url" mapstructure:"url"`
	ClientIDPrefix string           `json:"client_id_prefix" mapstructure:"client_id_prefix"`
	InstanceID     string           `json:"instance_id" mapstructure:"instance_id"`
	TopicPrefix    string           `json:"topic_prefix" mapstructure:"topic_prefix"`
	CleanSession   bool             `json:"clean_session" mapstructure:"clean_session"`
	KeepAlive      time.Duration    `json:"keep_alive" mapstructure:"keep_alive"`
	Reconnect      *ReconnectConfig `json:"reconnect" mapstructure:"reconnect"`
}

// ReconnectConfig 重连配置
type ReconnectConfig struct {
	Delay time.Duration `json:"delay" mapstructure:"delay"`
}

// Setup 初始化配置
func Setup(base ...string) *Config {
	cfg := &Config{
		Tracer: &TracerConfig{
			ServiceName: "aibuddy_hub",
		},
	}
	basePath := FoundConfigPath()

	basePath = filepath.ToSlash(basePath)
	if len(base) > 0 && base[0] != "" {
		basePath = base[0]
	}

	err := godotenv.Load(filepath.Join(basePath, "..", ".env"))
	if err != nil {
		slog.Debug("Could not load .env file", "error", err)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(basePath)

	viper.SetEnvPrefix("AIBUDDY_HUB")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err = viper.ReadInConfig()
	if err != nil {
		slog.Warn("加载配置文件失败，使用默认值和环境变量", "error", err)
	} else {
		slog.Info("加载配置文件成功", "file", viper.ConfigFileUsed())
	}

	for _, config := range []string{"dev", "prod", "test"} {
		configFile := filepath.ToSlash(filepath.Join(basePath, "config."+config+".yaml"))
		viper.SetConfigFile(configFile)
		if err = viper.MergeInConfig(); err != nil {
			slog.Warn("合并配置文件失败", "error", err)
		} else {
			slog.Info("合并配置文件成功", "file", viper.ConfigFileUsed())
		}
	}

	viper.AutomaticEnv()

	err = viper.Unmarshal(cfg)
	if err != nil {
		slog.Error("解析配置文件失败", "error", err)
		panic(err)
	}
	Instance = cfg
	return cfg
}

// FoundConfigPath 查找配置文件路径
func FoundConfigPath() string {
	// 1. 检查环境变量
	if path := checkEnvConfig(); path != "" {
		return path
	}

	// 2. 检查可执行文件同目录
	if path := checkExeConfig(); path != "" {
		return path
	}

	// 3. 检查调用者所在目录并向上查找
	if path := checkCallerConfig(); path != "" {
		return path
	}

	return defaultConfigPath()
}

func checkEnvConfig() string {
	if envPath := os.Getenv("AIBUDDY_CONFIG_PATH"); envPath != "" {
		if filepath.IsAbs(envPath) {
			return envPath
		}
		if abs, err := filepath.Abs(envPath); err == nil {
			return abs
		}
		return envPath
	}
	return ""
}

func checkExeConfig() string {
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		return checkConfigDir(dir)
	}
	return ""
}

func checkCallerConfig() string {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		return ""
	}

	dir := filepath.Dir(file)
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			configPath := filepath.Join(dir, "config")
			if _, err := os.Stat(configPath); err == nil {
				return configPath
			}
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func checkConfigDir(dir string) string {
	configPath := filepath.Join(dir, "config")
	if _, err := os.Stat(configPath); err == nil {
		return configPath
	}
	configPath = filepath.Join(dir, "..", "config")
	if _, err := os.Stat(configPath); err == nil {
		return filepath.Clean(configPath)
	}
	return ""
}

func defaultConfigPath() string {
	abs, err := filepath.Abs("./config")
	if err != nil {
		return "./config"
	}
	return abs
}
