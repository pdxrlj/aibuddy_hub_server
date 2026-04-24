// Package websocket 提供 websocket 服务
package websocket

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/olahol/melody"
	"github.com/spf13/cast"
)

// internalMessage 内部消息结构，用于在 channel 中传递
type internalMessage struct {
	UID   string
	Frame []byte
}

// SendChan 发送通道
var SendChan = make(chan internalMessage, 100)

// Service 提供 websocket 服务
type Service struct {
	Melody   *melody.Melody
	ConnPool *ConnPool
}

// NewWebsocket 创建 websocket 服务
func NewWebsocket() *Service {
	m := melody.New()

	s := &Service{
		Melody:   m,
		ConnPool: NewConnPool(),
	}

	// 设置连接回调
	m.HandleConnect(func(session *melody.Session) {
		uid, ok := session.Get("uid")
		if !ok {
			slog.Error("[Websocket] uid not found in session")
			return
		}
		uidStr := cast.ToString(uid)
		s.ConnPool.Add(uidStr, session)
		slog.Info("[Websocket] client connected", "uid", uidStr)
	})

	// 设置断开连接回调
	m.HandleDisconnect(func(session *melody.Session) {
		uid, ok := session.Get("uid")
		if !ok {
			slog.Error("uid is required")
			return
		}
		uidStr := cast.ToString(uid)
		slog.Info("[Websocket] HandleDisconnect", "uid", uidStr)
		s.ConnPool.Remove(uidStr)
	})

	// 设置消息回调
	m.HandleMessage(func(session *melody.Session, msg []byte) {
		s.HandleMessage(session, msg)
	})

	go s.StartMessageHandler() // 处理消息转发
	return s
}

// HandleConnect 处理 WebSocket 连接请求
func (s *Service) HandleConnect(uid int64, w http.ResponseWriter, r *http.Request) error {
	// 将 uid 存入 session
	keys := map[string]any{"uid": uid}
	if err := s.Melody.HandleRequestWithKeys(w, r, keys); err != nil {
		slog.Error("[Websocket] HandleRequestWithKeys error", "error", err)
		return err
	}
	return nil
}

// HandleMessage 处理消息
func (s *Service) HandleMessage(session *melody.Session, msg []byte) {
	// uid, ok := session.Get("uid")
	// if !ok {
	// 	slog.Error("uid is required")
	// 	return
	// }
	// slog.Info("[Websocket] HandleMessage", "uid", uid, "message", string(msg))

	var frame struct {
		Type     FrameType       `json:"type"`
		DeviceID string          `json:"device_id"`
		Message  json.RawMessage `json:"message"`
	}
	if err := json.Unmarshal(msg, &frame); err != nil {
		slog.Error("failed to decode frame", "error", err)
		return
	}

	switch frame.Type {
	case FrameTypePing:
		// 心跳消息，返回 pong
		pong := map[string]any{"type": "pong", "timestamp": time.Now().Unix()}
		if data, err := json.Marshal(pong); err == nil {
			if err := session.Write(data); err != nil {
				slog.Error("failed to write pong", "error", err)
			}
		}
	case FrameTypeUserMsg:
		// 收到用户消息，转发给设备
		// TODO: 实现转发逻辑
	default:
		slog.Warn("unsupported frame type", "type", frame.Type)
	}
}

// StartMessageHandler 启动消息处理协程
func (s *Service) StartMessageHandler() {
	for msg := range SendChan {
		session, ok := s.ConnPool.Get(msg.UID)
		if !ok {
			slog.Warn("[Websocket] user not online", "uid", msg.UID)
			continue
		}

		if err := session.Write(msg.Frame); err != nil {
			slog.Error("[Websocket] send message failed", "uid", msg.UID, "error", err)
		} else {
			slog.Debug("[Websocket] message sent", "uid", msg.UID)
		}
	}
}

// SendMessage 发送消息给指定用户
func SendMessage(uid string, frame Frame) {
	encodeFrame, err := frame.Encode()
	if err != nil {
		slog.Error("[Websocket] failed to encode frame", "error", err)
		return
	}

	select {
	case SendChan <- internalMessage{UID: uid, Frame: encodeFrame}:
	default:
		slog.Warn("[Websocket] send channel is full, message dropped", "uid", uid)
	}
}
