// Package repository provides the repository for the NFC card
package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
)

// NFCRepository NFC仓库
type NFCRepository struct {
}

// NewNFCRepository 创建NFC仓库
func NewNFCRepository() *NFCRepository {
	return &NFCRepository{}
}

// Create 创建NFC
func (r *NFCRepository) Create(nfc *model.NFC) error {
	return model.Conn().GetDB().Create(nfc).Error
}

// Update 更新NFC
func (r *NFCRepository) Update(nfc *model.NFC) error {
	return model.Conn().GetDB().Save(nfc).Error
}

// UpdateStatus 更新NFC状态
func (r *NFCRepository) UpdateStatus(cid string, status model.NFCStatus) error {
	_, err := query.NFC.Where(query.NFC.Cid.Eq(cid)).Update(query.NFC.Status, status)
	return err
}

// GetByCid 根据Cid获取NFC
func (r *NFCRepository) GetByCid(cid string) (*model.NFC, error) {
	nfc, err := query.NFC.Where(query.NFC.Cid.Eq(cid)).First()
	return nfc, err
}

// GetByNFCID 根据NFCID获取NFC
func (r *NFCRepository) GetByNFCID(nfcID string) (*model.NFC, error) {
	nfc, err := query.NFC.Where(query.NFC.NFCID.Eq(nfcID)).First()
	return nfc, err
}
