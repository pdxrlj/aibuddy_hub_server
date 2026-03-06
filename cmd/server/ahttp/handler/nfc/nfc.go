// Package nfc provides the handler for the NFC
package nfc

import (
	"aibuddy/internal/services/nfc"
	"aibuddy/pkg/ahttp"
	"net/http"
)

// Handler NFC处理器
type Handler struct {
	Service *nfc.Service
}

// NewHandler 创建NFC处理器
func NewHandler() *Handler {
	return &Handler{
		Service: nfc.NewNFC(),
	}
}

// CreateNFC 创建NFC
func (h *Handler) CreateNFC(state *ahttp.State, req *CreateNFCRequest) error {
	err := h.Service.CreateNFC(req.Ctype, req.Title, req.Content)
	if err != nil {
		return err
	}

	return state.Resposne().Success()
}

// GetNFCInfo 获取NFC信息
func (h *Handler) GetNFCInfo(state *ahttp.State, req *GetNFCInfoRequest) error {
	nfc, err := h.Service.GetNFCInfoByNFCID(req.NFCID)
	if err != nil {
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	return state.Resposne().Success(&GetNFCInfoResponse{
		NFCID:   nfc.NFCID,
		CID:     nfc.Cid,
		Ctype:   nfc.Ctype,
		Title:   nfc.Title,
		Content: nfc.Content,
	})
}
