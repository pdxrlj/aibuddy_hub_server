// Package nfc provides the service for the NFC
package nfc

import (
	nfcframe "aibuddy/aiframe/nfc"
	"aibuddy/internal/model"
	"aibuddy/internal/repository"

	"aibuddy/pkg/helpers"
)

// Service NFC服务
type Service struct {
	nfcRepository *repository.NFCRepository
}

// NewNFC 创建NFC服务
func NewNFC() *Service {
	return &Service{
		nfcRepository: repository.NewNFCRepository(),
	}
}

// CreateNFC 创建NFC
func (s *Service) CreateNFC(ctype string, title, content string) error {
	cid := helpers.GenerateNumber(10)
	nfc := &model.NFC{
		Cid:     cid,
		Ctype:   ctype,
		Title:   title,
		Content: content,
	}
	if err := s.nfcRepository.Create(nfc); err != nil {
		return err
	}

	// 发送MQTT信息
	if err := nfcframe.SendNFCCreate(nfc.DeviceID, cid, ctype); err != nil {
		return err
	}

	return nil
}

// GetNFCInfoByNFCID 根据NFCID获取NFC信息
func (s *Service) GetNFCInfoByNFCID(nfcID string) (*model.NFC, error) {
	return s.nfcRepository.GetByNFCID(nfcID)
}
