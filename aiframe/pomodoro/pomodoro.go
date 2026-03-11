// Package pomodoro 番茄钟
package pomodoro

import "encoding/json"

// Pomodoro is the pomodoro timer
type Pomodoro struct {
	Type                string `json:"type"`                 // pomodoro type
	TotalTime           int    `json:"total_time"`           // 总时间，单位：分钟
	StudyDuration       int    `json:"study_duration"`       // 学习时间，单位：分钟
	DistractionDuration int    `json:"distraction_duration"` // 分心时间，单位：分钟
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
