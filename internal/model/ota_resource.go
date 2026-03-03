// Package model provides the ota resource model.
package model

import "time"

// OtaResource OTA 资源
type OtaResource struct {
	ID int `gorm:"column:id;type:int(11);primaryKey;autoIncrement;"`

	OtaURL      string `gorm:"column:ota_url;type:varchar(255);not null;"`
	ModelURL    string `gorm:"column:model_url;type:varchar(255);not null;"`
	ResourceURL string `gorm:"column:resource_url;type:varchar(255);not null;"`
	ForceUpdate bool   `gorm:"column:force_update;type:boolean;not null;"`

	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;not null;"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;not null;"`
}

// TableName 表名
func (OtaResource) TableName() string {
	return TableName("ota_resource")
}
