// Package pomodoro 番茄钟
package pomodoro

import "encoding/json"

// DistractionRecord 分心记录
type DistractionRecord struct {
	TouchScreenCount int `json:"touch_screen_count"` // 触摸屏幕次数
	TouchHead        int `json:"touch_head"`         // 触摸头部次数
}

// Pomodoro is the pomodoro timer
type Pomodoro struct {
	Type              string            `json:"type"`               // pomodoro type
	TotalTime         int               `json:"total_time"`         // 总时间，单位：秒
	StudyDuration     int               `json:"study_duration"`     // 学习时间，单位：秒
	DistractionRecord *DistractionRecord `json:"distraction_record"` // 分心记录
}

// NewPomodoro creates a new pomodoro timer
func NewPomodoro() *Pomodoro {
	return &Pomodoro{}
}

// Encode encodes the pomodoro to json
func (p *Pomodoro) Encode() ([]byte, error) {
	return json.Marshal(p)
}

// Decode decodes the pomodoro from json
func (p *Pomodoro) Decode(data []byte) error {
	return json.Unmarshal(data, p)
}
