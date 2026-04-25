// Package model provides the ota resource model.
package model

import "time"

// OtaResource OTA 资源
type OtaResource struct {
	ID int `gorm:"column:id;type:int(11);primaryKey;autoIncrement;"`

	Version     string `gorm:"column:version;type:varchar(255);not null;comment:版本号;"`
	OtaURL      string `gorm:"column:ota_url;type:varchar(255);not null;comment:OTA下载地址;"`
	ModelURL    string `gorm:"column:model_url;type:varchar(255);not null;comment:模型下载地址;"`
	ResourceURL string `gorm:"column:resource_url;type:varchar(255);not null;comment:资源下载地址;"`
	ForceUpdate bool   `gorm:"column:force_update;type:boolean;not null;default:false;comment:是否强制更新;"`

	BoardType string `gorm:"column:board_type;type:varchar(255);default:nl;not null;comment:板型;"`

	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;not null;"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;not null;"`
}

// TableName 表名
func (OtaResource) TableName() string {
	return TableName("ota_resource")
}
