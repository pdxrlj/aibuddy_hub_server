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
func (s *Service) CreateNFC(uid int64, deviceID, ctype string, title, content string) error {
	cid := helpers.GenerateNumber(10)
	nfc := &model.NFC{
		UID:      uid,
		DeviceID: deviceID,
		Cid:      cid,
		Ctype:    ctype,
		Title:    title,
		Content:  content,
	}

	if err := s.nfcRepository.Create(nfc); err != nil {
		return err
	}

	// 发送MQTT信息
	if err := nfcframe.SendNFCCreate(deviceID, cid, ctype); err != nil {
		return err
	}

	return nil
}

// GetNFCInfoByNFCID 根据NFCID获取NFC信息
func (s *Service) GetNFCInfoByNFCID(nfcID string) (*model.NFC, error) {
	return s.nfcRepository.GetByNFCID(nfcID)
}

// GetNFCListByDeviceID 根据设备ID获取NFC列表
func (s *Service) GetNFCListByDeviceID(deviceID string, page, pageSize int) ([]*model.NFC, int64, error) {
	return s.nfcRepository.GetListByDeviceID(deviceID, page, pageSize)
}

// UpdateNFC 更新NFC
func (s *Service) UpdateNFC(cid, ctype, title, content string) error {
	nfc, err := s.nfcRepository.GetByCid(cid)
	if err != nil {
		return err
	}

	nfc.Ctype = ctype
	nfc.Title = title
	nfc.Content = content

	return s.nfcRepository.Update(nfc)
}

// DeleteNFC 删除NFC
func (s *Service) DeleteNFC(cid string) error {
	return s.nfcRepository.Delete(cid)
}
