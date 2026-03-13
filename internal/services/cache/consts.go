// Package cache 提供缓存服务
package cache

import (
	"fmt"

	"github.com/spf13/cast"
)

const (
	// RTCInstanceID 实例ID缓存Key模板
	RTCInstanceID = "rtc_instance_id:%s"
)

// StoreRTCInstanceID 存储实例ID
func StoreRTCInstanceID(appID string, instanceID string) error {
	return flashInstance.Set(fmt.Sprintf(RTCInstanceID, appID), instanceID)
}

// GetRTCInstanceID 获取实例ID
func GetRTCInstanceID(appID string) (string, error) {
	instanceID, err := flashInstance.Get(fmt.Sprintf(RTCInstanceID, appID))
	if err != nil {
		return "", err
	}
	return cast.ToString(instanceID), nil
}
