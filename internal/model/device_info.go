package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
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
	ID          int64     `gorm:"primaryKey;type:bigint(20);autoIncrement;column:id;"`
	DeviceID    string    `gorm:"column:device_id;index;type:varchar(32);not null;uniqueIndex;comment:设备ID;"`
	NickName    string    `gorm:"column:nickname;type:varchar(16);not null;comment:设备昵称;"`
	Avatar      string    `gorm:"column:avatar;type:varchar(255);comment:头像;"`
	Gender      int8      `gorm:"column:gender;type:int;default:0;oneof=0 1 2;comment:性别;"`
	Birthday    time.Time `gorm:"column:birthday;type:date;not null;comment:生日;"`
	Relation    string    `gorm:"column:relation;type:varchar(8);not null;comment:关系;"`
	Hobbies     JSONArray `gorm:"column:hobbies;type:json;not null;comment:兴趣;"`
	Values      JSONArray `gorm:"column:values;type:json;not null;comment:价值观;"`
	Skills      JSONArray `gorm:"column:skills;type:json;not null;comment:技能;"`
	Personality JSONArray `gorm:"column:personality;type:json;not null;comment:性格;"`
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamp;not null;comment:创建时间;"`
}

// TableName 表名
func (DeviceInfo) TableName() string {
	return TableName("device_info")
}
