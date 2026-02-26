// Package mqtt 提供MQTT客户端连接和管理功能
package mqtt

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"strings"
)

// AliyunClientIDSeparator 阿里云客户端ID分隔符
const AliyunClientIDSeparator = "@@@"

// AliyunClientIDInfo 阿里云客户端ID信息
type AliyunClientIDInfo struct {
	GroupID    string `json:"group_id"`    // 组ID
	DeviceID   string `json:"device_id"`   // 设备ID
	OriginalID string `json:"original_id"` // 原始完整ID
}

// IsAliyunClientID 判断客户端ID是否为阿里云格式
func IsAliyunClientID(clientID string) bool {
	return strings.Contains(clientID, AliyunClientIDSeparator)
}

// ValidateAliyunClientIDFormat 验证客户端ID格式
func ValidateAliyunClientIDFormat(clientID string) error {
	if !IsAliyunClientID(clientID) {
		return fmt.Errorf("客户端ID不是阿里云格式: %s", clientID)
	}

	parts := strings.Split(clientID, AliyunClientIDSeparator)
	if len(parts) != 2 {
		return fmt.Errorf("阿里云客户端ID格式错误，应为 groupid@@@deviceid: %s", clientID)
	}

	groupID := parts[0]
	deviceID := parts[1]

	// 验证组ID
	if groupID == "" {
		return fmt.Errorf("组ID不能为空")
	}

	// 验证设备ID
	if deviceID == "" {
		return fmt.Errorf("设备ID不能为空")
	}

	// 检查是否包含非法字符
	if strings.ContainsAny(groupID, " \t\n\r") {
		return fmt.Errorf("组ID包含非法字符")
	}
	if strings.ContainsAny(deviceID, " \t\n\r") {
		return fmt.Errorf("设备ID包含非法字符")
	}

	return nil
}

// BuildAliyunMQTTUsername 构建阿里云MQTT用户名（签名模式）
// 规范：Signature|AccessKeyId|InstanceId
func BuildAliyunMQTTUsername(accessKeyID, instanceID string) string {
	return fmt.Sprintf("%s|%s|%s", "Signature", accessKeyID, instanceID)
}

// BuildAliyunMQTTPassword 构建阿里云MQTT密码（签名模式）
func BuildAliyunMQTTPassword(clientID, accessKeySecret string) (string, error) {
	if clientID == "" {
		return "", fmt.Errorf("clientID 不能为空")
	}
	if accessKeySecret == "" {
		return "", fmt.Errorf("accessKeySecret 不能为空")
	}

	mac := hmac.New(sha1.New, []byte(accessKeySecret))
	if _, err := mac.Write([]byte(clientID)); err != nil {
		return "", fmt.Errorf("签名计算失败: %w", err)
	}
	sig := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(sig), nil
}

// GenerateAliyunMQTTAuth 生成阿里云MQTT认证信息
// clientID: 客户端ID
// accessKeyID: 阿里云访问密钥ID
// accessKeySecret: 阿里云访问密钥密钥
// instanceID: 阿里云实例ID
func GenerateAliyunMQTTAuth(clientID, accessKeyID, accessKeySecret, instanceID string) (string, string, error) {
	if err := ValidateAliyunClientIDFormat(clientID); err != nil {
		return "", "", fmt.Errorf("客户端ID格式错误: %w", err)
	}

	if accessKeyID == "" || accessKeySecret == "" || instanceID == "" {
		return "", "", fmt.Errorf("缺少必要的阿里云凭据: accessKeyId/accessKeySecret/instanceId")
	}

	username := BuildAliyunMQTTUsername(accessKeyID, instanceID)
	password, err := BuildAliyunMQTTPassword(clientID, accessKeySecret)
	if err != nil {
		return "", "", err
	}
	return username, password, nil
}
