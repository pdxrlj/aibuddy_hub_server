// Package websocket 提供 websocket 帧处理
package websocket

import "encoding/json"

// FrameType 帧类型
type FrameType string

const (
	// FrameTypeOnline 设备在线状态帧
	FrameTypeOnline FrameType = "online"
	// FrameTypeOffline 设备离线状态帧
	FrameTypeOffline FrameType = "offline"
	// FrameTypeDeviceMsg 设备消息帧
	FrameTypeDeviceMsg FrameType = "device_message"
	// FrameTypeUserMsg 用户消息帧
	FrameTypeUserMsg FrameType = "user_message"
	// FrameTypeNFCCreateSuccess NFC创建成功帧
	FrameTypeNFCCreateSuccess FrameType = "nfc_create_success"
)

// Frame 帧接口
type Frame interface {
	Encode() ([]byte, error)
	Decode(data []byte) error
}

// DeviceOnlineFrame 设备在线状态帧
type DeviceOnlineFrame struct {
	Type     FrameType       `json:"type"`
	DeviceID string          `json:"device_id"`
	Message  json.RawMessage `json:"message"`
}

// DeviceOfflineFrame 设备离线状态帧
type DeviceOfflineFrame struct {
	Type     FrameType       `json:"type"`
	DeviceID string          `json:"device_id"`
	Message  json.RawMessage `json:"message"`
}

// DeviceToUserFrame 设备消息帧
type DeviceToUserFrame struct {
	Type     FrameType       `json:"type"`
	DeviceID string          `json:"device_id"`
	Message  json.RawMessage `json:"message"`
}

// UserToDeviceFrame 用户到设备帧
type UserToDeviceFrame struct {
	Type     FrameType       `json:"type"`
	DeviceID string          `json:"device_id"`
	Message  json.RawMessage `json:"message"`
}

// Decode 解码设备消息帧
func (f *DeviceToUserFrame) Decode(data []byte) error {
	return json.Unmarshal(data, f)
}

// Decode 解码用户消息帧
func (f *UserToDeviceFrame) Decode(data []byte) error {
	return json.Unmarshal(data, f)
}

// Decode 解码设备在线帧
func (f *DeviceOnlineFrame) Decode(data []byte) error {
	return json.Unmarshal(data, f)
}

// Encode 编码设备在线帧
func (f *DeviceOnlineFrame) Encode() ([]byte, error) {
	return json.Marshal(f)
}

// Decode 解码设备离线帧
func (f *DeviceOfflineFrame) Decode(data []byte) error {
	return json.Unmarshal(data, f)
}

// Encode 编码设备离线帧
func (f *DeviceOfflineFrame) Encode() ([]byte, error) {
	return json.Marshal(f)
}

// Encode 编码设备消息帧
func (f *DeviceToUserFrame) Encode() ([]byte, error) {
	return json.Marshal(f)
}

// Encode 编码用户到设备帧
func (f *UserToDeviceFrame) Encode() ([]byte, error) {
	return json.Marshal(f)
}

// NFCCreateSuccessFrame 创建成功帧
type NFCCreateSuccessFrame struct {
	Type     FrameType `json:"type"`
	CID      string    `json:"cid"`
	DeviceID string    `json:"device_id"`
	NFCID    string    `json:"nfc_id"`
}

// Encode 编码NFC创建成功帧
func (f *NFCCreateSuccessFrame) Encode() ([]byte, error) {
	return json.Marshal(f)
}

// Decode 解码NFC创建成功帧
func (f *NFCCreateSuccessFrame) Decode(data []byte) error {
	return json.Unmarshal(data, f)
}
