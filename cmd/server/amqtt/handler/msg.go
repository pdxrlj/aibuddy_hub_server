package handler

import (
	"aibuddy/aiframe/message"
	"aibuddy/pkg/mqtt"
	"fmt"
	"log/slog"
)

// MsgHandler 消息处理器
type MsgHandler struct{}

// NewMsgHandler 创建消息处理器
func NewMsgHandler() *MsgHandler {
	return &MsgHandler{}
}

// Handle 处理消息
func (h *MsgHandler) Handle(ctx *mqtt.Context) {
	defer ctx.Message.Ack()

	deviceID := ctx.Params["device_id"]
	var baseMsg message.Msg
	if err := ctx.BindJSON(&baseMsg); err != nil {
		ctx.Message.Ack()
		slog.Error("[MQTT] BindJSON failed", "error", err)
		return
	}

	_ = deviceID

	if err := h.decodeMessage(ctx, baseMsg.Type); err != nil {
		slog.Error("decodeMessage failed", "error", err)
	}
}

// msgParsers 消息类型解析器映射
var msgParsers = map[message.MsgType]func(ctx *mqtt.Context) error{
	message.MsgTypeFriendsReq: func(ctx *mqtt.Context) error {
		var m message.FriendsReqMsg
		return ctx.BindJSON(&m)
	},
	message.MsgTypeFriends: func(ctx *mqtt.Context) error {
		var m message.FriendsMsg
		return ctx.BindJSON(&m)
	},
	message.MsgTypeSend: func(ctx *mqtt.Context) error {
		var m message.SendMsg
		return ctx.BindJSON(&m)
	},
	message.MsgTypeRecv: func(ctx *mqtt.Context) error {
		var m message.RecvMsg
		return ctx.BindJSON(&m)
	},
	message.MsgTypeRead: func(ctx *mqtt.Context) error {
		var m message.MarkReadMsg
		return ctx.BindJSON(&m)
	},
	message.MsgTypeUnreadReq: func(ctx *mqtt.Context) error {
		var m message.UnreadReqMsg
		return ctx.BindJSON(&m)
	},
	message.MsgTypeUnread: func(ctx *mqtt.Context) error {
		var m message.UnreadMsg
		return ctx.BindJSON(&m)
	},
	message.MsgTypeSendFail: func(ctx *mqtt.Context) error {
		var m message.SendFailMsg
		return ctx.BindJSON(&m)
	},
}

// decodeMessage 根据数据类型解码消息
func (h *MsgHandler) decodeMessage(ctx *mqtt.Context, msgType message.MsgType) error {
	parser, ok := msgParsers[msgType]
	if !ok {
		return fmt.Errorf("unknown message type: %s", msgType)
	}
	return parser(ctx)
}
