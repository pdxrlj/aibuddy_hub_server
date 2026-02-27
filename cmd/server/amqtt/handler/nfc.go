// Package handler 定义NFC相关的消息处理器
package handler

import (
	"aibuddy/aiframe/nfc"
	logger "aibuddy/pkg/log"
	"aibuddy/pkg/mqtt"
	"fmt"
	"log/slog"
)

// NFCHandler NFC消息处理器
type NFCHandler struct{}

// NewNFCHandler 创建NFC消息处理器
func NewNFCHandler() *NFCHandler {
	return &NFCHandler{}
}

// Handle 处理NFC消息
func (h *NFCHandler) Handle(ctx *mqtt.Context) {
	defer ctx.Message.Ack()

	deviceID := ctx.Params["device_id"]
	var baseMsg nfc.BaseMsg
	if err := ctx.BindJSON(&baseMsg); err != nil {
		slog.Error(logger.MQTTNFC, "error", err, "msg", "bind json failed")
		return
	}

	if err := h.handleNFCMessage(deviceID, baseMsg.Type, ctx); err != nil {
		slog.Error(logger.MQTTNFC, "error", err, "msg", "handle nfc message failed")
	}
}

func (h *NFCHandler) handleNFCMessage(deviceID string, msgType nfc.Type, ctx *mqtt.Context) error {
	switch msgType {
	case nfc.TypeFind:
		return h.handleNFCFind(deviceID, ctx)
	case nfc.TypeFindRes:
		return h.handleNFCFindRes(deviceID, ctx)
	case nfc.TypeAddReq:
		return h.handleNFCAddReq(deviceID, ctx)
	case nfc.TypeDel:
		return h.handleNFCDel(deviceID, ctx)
	case nfc.TypeCreated:
		return h.handleNFCCreated(deviceID, ctx)
	case nfc.TypeInfoReq:
		return h.handleNFCInfoReq(deviceID, ctx)
	default:
		return fmt.Errorf("unknown nfc message type: %s", msgType)
	}
}

func (h *NFCHandler) handleNFCFind(deviceID string, ctx *mqtt.Context) error {
	var msg nfc.FindMsg
	if err := ctx.BindJSON(&msg); err != nil {
		return err
	}
	// 业务逻辑处理
	slog.Info("NFC Find", "device", deviceID, "msg", msg)
	return nil
}

func (h *NFCHandler) handleNFCFindRes(deviceID string, ctx *mqtt.Context) error {
	var msg nfc.FindResMsg
	if err := ctx.BindJSON(&msg); err != nil {
		return err
	}
	// 业务逻辑处理
	slog.Info("NFC FindRes", "device", deviceID, "msg", msg)
	return nil
}

func (h *NFCHandler) handleNFCAddReq(deviceID string, ctx *mqtt.Context) error {
	var msg nfc.AddReqMsg
	if err := ctx.BindJSON(&msg); err != nil {
		return err
	}
	// 业务逻辑处理
	slog.Info("NFC AddReq", "device", deviceID, "msg", msg)
	return nil
}

func (h *NFCHandler) handleNFCDel(deviceID string, ctx *mqtt.Context) error {
	var msg nfc.DelMsg
	if err := ctx.BindJSON(&msg); err != nil {
		return err
	}
	// 业务逻辑处理
	slog.Info("NFC Del", "device", deviceID, "msg", msg)
	return nil
}

func (h *NFCHandler) handleNFCCreated(deviceID string, ctx *mqtt.Context) error {
	var msg nfc.CreatedMsg
	if err := ctx.BindJSON(&msg); err != nil {
		return err
	}
	// 业务逻辑处理
	slog.Info("NFC Created", "device", deviceID, "msg", msg)
	return nil
}

func (h *NFCHandler) handleNFCInfoReq(deviceID string, ctx *mqtt.Context) error {
	var msg nfc.InfoReqMsg
	if err := ctx.BindJSON(&msg); err != nil {
		return err
	}
	// 业务逻辑处理
	slog.Info("NFC InfoReq", "device", deviceID, "msg", msg)
	return nil
}
