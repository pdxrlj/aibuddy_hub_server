// Package cache 提供缓存服务
package cache

import (
	"fmt"
	"strings"

	"github.com/spf13/cast"
)

const (
	// RTCInstanceID 实例ID缓存Key模板
	RTCInstanceID = "rtc_instance_id:%s"
)

// StoreRTCInstanceID 存储实例ID
func StoreRTCInstanceID(deviceID string, instanceID string) error {
	deviceID = strings.ReplaceAll(deviceID, ":", "-")

	return flashInstance.Set(fmt.Sprintf(RTCInstanceID, deviceID), instanceID)
}

// GetRTCInstanceID 获取实例ID
func GetRTCInstanceID(deviceID string) (string, error) {
	deviceID = strings.ReplaceAll(deviceID, ":", "-")
	instanceID, err := flashInstance.Get(fmt.Sprintf(RTCInstanceID, deviceID))
	if err != nil {
		return "", err
	}
	return cast.ToString(instanceID), nil
}
