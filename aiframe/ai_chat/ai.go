// Package ai 实现了 AI 对话相关功能
package ai

import "encoding/json"

// AI 对话
// 1. 对话开始（设备 → 云端）
// Topic: aibuddy/{device_id}/ai

// {
//     "type": "start",
//     "sid": "nailong_1679036400_123"
// }
// 2. 对话结束（设备 → 云端）
// Topic: aibuddy/{device_id}/ai

// {
//     "type": "end",
//     "sid": "nailong_1679036400_123",
//     "dur": 120
// }
// 3. 角色切换（设备 → 云端）
// Topic: aibuddy/{device_id}/ai

// {
//     "type": "role",
//     "role": "nailong"
// }

// ChatType AI 对话类型
type ChatType string

// ChatTypeStart 对话开始
const (
	// ChatTypeStart 对话开始
	ChatTypeStart ChatType = "start"
	// ChatTypeEnd 对话结束
	ChatTypeEnd ChatType = "end"
	// ChatTypeSwitchRole 角色切换
	ChatTypeSwitchRole ChatType = "switch_role"
)

// IsValid 验证 AI 对话类型是否有效
func (t ChatType) IsValid() bool {
	return t == ChatTypeStart || t == ChatTypeEnd || t == ChatTypeSwitchRole
}

// Chat AI 对话
type Chat struct {
	Type ChatType `json:"type,omitempty"`
	Sid  string   `json:"sid,omitempty"`
	Dur  int      `json:"dur,omitempty"`
	Role string   `json:"role,omitempty"`
}

// Encode 编码 AI 对话
func (a *Chat) Encode() ([]byte, error) {
	return json.Marshal(a)
}

// Decode 解码 AI 对话
func (a *Chat) Decode(data []byte) error {
	return json.Unmarshal(data, a)
}
