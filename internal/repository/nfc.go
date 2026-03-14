// Package repository provides the repository for the NFC card
package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"time"
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

// GetListByDeviceID 根据设备ID获取NFC列表（分页）
func (r *NFCRepository) GetListByDeviceID(deviceID string, page, pageSize int) ([]*model.NFC, int64, error) {
	list, total, err := query.NFC.Where(query.NFC.DeviceID.Eq(deviceID)).FindByPage((page-1)*pageSize, pageSize)
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// Delete 删除NFC
func (r *NFCRepository) Delete(cid string) error {
	_, err := query.NFC.Where(query.NFC.Cid.Eq(cid)).Delete()
	return err
}

// GetNfcDataByDeviceID 获取设备在指定时间范围内的NFC数据
func (r *NFCRepository) GetNfcDataByDeviceID(deviceID string, startTime, endTime time.Time) ([]*model.NFC, error) {
	nfcData, err := query.NFC.Where(query.NFC.DeviceID.Eq(deviceID)).Where(query.NFC.CreatedAt.Between(startTime, endTime)).Find()
	if err != nil {
		return nil, err
	}

	return nfcData, nil
}
