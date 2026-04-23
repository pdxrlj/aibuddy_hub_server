// Package handler 定义NFC相关的消息处理器
package handler

import (
	"aibuddy/aiframe/nfc"
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"aibuddy/internal/repository"
	"aibuddy/internal/services/websocket"
	"aibuddy/pkg/mqtt"
	"context"
	"errors"
	"log/slog"

	"github.com/spf13/cast"
)

// NFCHandler NFC消息处理器
type NFCHandler struct {
	nfcRepository    *repository.NFCRepository
	deviceRepository *repository.DeviceRepo
}

// NewNFCHandler 创建NFC消息处理器
func NewNFCHandler() *NFCHandler {
	return &NFCHandler{
		nfcRepository:    repository.NewNFCRepository(),
		deviceRepository: repository.NewDeviceRepo(),
	}
}

// Handle 处理NFC消息
func (h *NFCHandler) Handle(ctx *mqtt.Context) {
	defer ctx.Message.Ack()

	deviceID := ctx.Params["device_id"]
	var msg nfc.CreateResRequest

	if err := msg.Decode(ctx.Payload); err != nil {
		slog.Error("[MQTT] NFC", "device_id", deviceID, "error", err)
		return
	}

	if msg.NFCID == "" {
		slog.Warn("[MQTT] NFC", "device_id", deviceID, "error", errors.New("server info nfc id is null"))
		return
	}

	nfc, err := h.nfcRepository.GetByCid(msg.Cid)
	if err != nil {
		slog.Error("[MQTT] NFC", "device_id", deviceID, "error", err)
		return
	}

	if nfc == nil {
		slog.Error("[MQTT] NFC", "device_id", deviceID, "error", "NFC not found")
		return
	}

	if nfc.Status != model.NFCPending {
		slog.Error("[MQTT] NFC", "device_id", deviceID, "error", "NFC already created")
		return
	}

	nfc.Status = model.NFCPaid
	nfc.NFCID = msg.NFCID
	nfc.DeviceID = deviceID

	if err := query.Q.Transaction(func(tx *query.Query) error {
		// 之前的设置为失效
		if err := h.nfcRepository.UpdateNFCStatusInvalidByNFCID(nfc.NFCID, tx); err != nil {
			return err
		}

		slog.Info("[NFC]", "event", "收到制作完成通知")

		return h.nfcRepository.Update(nfc, tx)
	}); err != nil {
		slog.Error("[MQTT] NFC", "device_id", deviceID, "error", err)
		return
	}

	slog.Info("[MQTT] NFC", "device_id", deviceID, "nfc_id", nfc.NFCID)

	// 发送NFC消息
	device, err := h.deviceRepository.GetDeviceInfo(context.Background(), deviceID)
	if err != nil {
		slog.Error("[MQTT] NFC", "device_id", deviceID, "found device error", err, "nfc_id", nfc.NFCID)
		return
	}

	websocket.SendMessage(cast.ToString(device.UID), &websocket.NFCCreateSuccessFrame{
		Type:     websocket.FrameTypeNFCCreateSuccess,
		CID:      nfc.Cid,
		DeviceID: deviceID,
		NFCID:    nfc.NFCID,
	})
}
