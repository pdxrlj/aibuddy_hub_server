package model

import (
	"aibuddy/pkg/config"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// JSONArray 自定义JSON数组类型，用于处理数组字段
type JSONArray []string

// Value 实现driver.Valuer接口
func (j JSONArray) Value() (driver.Value, error) {
	if len(j) == 0 {
		return "[]", nil
	}
	return json.Marshal(j)
}

// Scan 实现sql.Scanner接口
func (j *JSONArray) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONArray, 0)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("cannot scan non-[]byte value into JSONArray")
	}

	return json.Unmarshal(bytes, j)
}

// DeviceInfo 详细信息表结构
type DeviceInfo struct {
	ID       int64     `gorm:"primaryKey;type:bigint(20);autoIncrement;column:id;"`
	DeviceID string    `gorm:"column:device_id;index;type:varchar(32);not null;uniqueIndex;comment:设备ID;"`
	NickName string    `gorm:"column:nickname;type:varchar(16);not null;comment:设备昵称;"`
	Avatar   string    `gorm:"column:avatar;type:varchar(255);comment:头像;"`
	Gender   string    `gorm:"column:gender;type:varchar(16);default:'未知';comment:性别 未知 男性 女性;"`
	Birthday time.Time `gorm:"column:birthday;type:date;not null;comment:生日;"`

	Hobbies     JSONArray `gorm:"column:hobbies;type:json;not null;comment:兴趣;"`
	Values      JSONArray `gorm:"column:values;type:json;not null;comment:价值观;"`
	Skills      JSONArray `gorm:"column:skills;type:json;not null;comment:技能;"`
	Personality JSONArray `gorm:"column:personality;type:json;not null;comment:性格;"`
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamp;not null;comment:创建时间;"`

	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;not null;comment:更新时间;"`
}

// TableName 表名
func (DeviceInfo) TableName() string {
	return TableName("device_info")
}

// AfterFind 在查询到设备信息后，将头像URL转换为完整的URL
func (d *DeviceInfo) AfterFind(_ *gorm.DB) error {
	domainname := DefaultDomainName
	if config.Instance != nil && config.Instance.App != nil && config.Instance.App.DomainName != "" {
		domainname = config.Instance.App.DomainName
	}

	if d.Avatar != "" {
		d.Avatar = fmt.Sprintf("%s/api/v1/file/file_proxy?filename=%s", domainname, d.Avatar)
	}
	return nil
}

// BeforeCreate 在插入之前
func (d *DeviceInfo) BeforeCreate(_ *gorm.DB) (err error) {
	if d.DeviceID != "" {
		d.DeviceID = strings.ToUpper(d.DeviceID)
	}
	return nil
}

// BeforeUpdate 在更新之前,将DeviceID转换为大写
func (d *DeviceInfo) BeforeUpdate(_ *gorm.DB) (err error) {
	d.DeviceID = strings.ToUpper(d.DeviceID)
	if d.Avatar != "" {
		d.Avatar = ExtractFilename(d.Avatar)
	}
	return nil
}
