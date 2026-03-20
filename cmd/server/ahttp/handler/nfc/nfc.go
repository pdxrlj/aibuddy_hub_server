// Package nfc provides the handler for the NFC
package nfc

import (
	aiuserService "aibuddy/internal/services/aiuser"
	"aibuddy/internal/services/nfc"
	"aibuddy/pkg/ahttp"
	"fmt"
	"log/slog"
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
	uid, err := aiuserService.GetUIDFromContext(state.Ctx)
	if err != nil {
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}
	slog.Info("[NFC] CreateNFC", "uid", uid, "device_id", req.DeviceID, "ctype", req.Ctype, "title", req.Title, "content", req.Content)
	err = h.Service.CreateNFC(uid, req.DeviceID, req.Ctype, req.Title, req.Content, req.Voice, req.Picture, *req.Dur)
	if err != nil {
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
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

// GetNFCList 获取NFC列表
func (h *Handler) GetNFCList(state *ahttp.State, req *GetNFCListRequest) error {
	dur := 0
	if req.Dur != nil {
		dur = *req.Dur
	}

	fmt.Println(dur)
	list, total, err := h.Service.GetNFCListByDeviceID(req.DeviceID, req.UpdateAt, dur, req.Ctype, req.Page, req.PageSize)
	if err != nil {
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	items := make([]ListItem, len(list))
	for i, item := range list {
		items[i] = ListItem{
			CID:     item.Cid,
			Ctype:   item.Ctype,
			Title:   item.Title,
			Content: item.Content,
			Voice:   item.Voice,
			Picture: item.Picture,
			Status:  string(item.Status),
			Dur:     item.Dur,
		}
	}

	return state.Resposne().Success(&GetNFCListResponse{
		Page:     req.Page,
		PageSize: req.PageSize,
		Total:    total,
		List:     items,
	})
}

// UpdateNFC 更新NFC
func (h *Handler) UpdateNFC(state *ahttp.State, req *UpdateNFCRequest) error {
	err := h.Service.UpdateNFC(req.CID, req.Ctype, req.Title, req.Content, req.Voice, req.Picture, *req.Dur)
	if err != nil {
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	return state.Resposne().Success()
}

// DeleteNFC 删除NFC
func (h *Handler) DeleteNFC(state *ahttp.State, req *DeleteNFCRequest) error {
	err := h.Service.DeleteNFC(req.CID)
	if err != nil {
		return state.Resposne().SetStatus(http.StatusBadRequest).Error(err)
	}

	return state.Resposne().Success()
}
