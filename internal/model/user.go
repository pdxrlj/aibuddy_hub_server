package model

import "time"

// User represents a user in the system.
type User struct {
	ID       int64  `gorm:"primaryKey;autoIncrement;column:id;"`
	OpenID   string `gorm:"column:open_id;type:varchar(255);not null;uniqueIndex;"`
	Nickname string `gorm:"column:nickname;type:varchar(255);not null;"`
	Phone    string `gorm:"column:phone;index;type:varchar(255);not null;"`

	Avatar   string `gorm:"column:avatar;type:varchar(255);not null;"`
	ParentID int64  `gorm:"column:parent_id;index;type:bigint;not null;comment:推荐人ID;"`

	DeviceID string `gorm:"column:device_id;index;type:varchar(255);not null;"`

	Device *Device `gorm:"foreignKey:DeviceID;references:DeviceID;"`

	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;not null;"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;not null;"`
}

// TableName returns the table name for User model.
func (User) TableName() string {
	return TableName("user")
}
