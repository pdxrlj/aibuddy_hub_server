// Package aiframe 提供 AI 设备框架接口定义
package aiframe

// Frame 帧编码/解码接口
type Frame interface {
	Encode() ([]byte, error)
	Decode(data []byte) error
}
