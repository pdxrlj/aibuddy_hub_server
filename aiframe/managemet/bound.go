// Package management provides a management service for the device.
package management

import (
	"aibuddy/aiframe"
	"aibuddy/pkg/mqtt"
)

// SendBoundToDevice 发送绑定信息到设备
func (m *Mgmt) SendBoundToDevice(deviceID string) error {
	payload, err := m.Encode()
	if err != nil {
		return err
	}

	return mqtt.Instance.Publish(aiframe.MQTTBoundTopic(deviceID), 1, false, payload)
}

// SendUnboundToDevice 发送解绑信息到设备
func SendUnboundToDevice(deviceID string) error {
	unbound := &Mgmt{
		Type: MgmtTypeUnbind,
	}
	payload, err := unbound.Encode()
	if err != nil {
		return err
	}

	return mqtt.Instance.Publish(aiframe.MQTTUnbindTopic(deviceID), 1, false, payload)
}
