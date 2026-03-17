package model

import (
	"fmt"
	"time"
)

// User represents a user in the system.
type User struct {
	ID       int64  `gorm:"primaryKey;autoIncrement;column:id;"`
	OpenID   string `gorm:"column:open_id;type:varchar(255);not null;index;"`
	Nickname string `gorm:"column:nickname;type:varchar(255);not null;"`
	Phone    string `gorm:"column:phone;index;type:varchar(255);not null;"`

	Avatar string `gorm:"column:avatar;type:varchar(255);not null;"`

	Email    string    `gorm:"column:email;type:varchar(128);"`
	Username string    `gorm:"column:username;type:varchar(32);"`
	Gender   int       `gorm:"sex:gender;type:int;default:0"`
	Birthday time.Time `gorm:"column:birthday;type:date;comment:生日;"`

	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;not null;"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;not null;"`
}

// TableName returns the table name for User model.
func (User) TableName() string {
	return TableName("user")
}

func (u *User) String() string {
	return fmt.Sprintf("ID: %d, OpenID: %s, Nickname: %s, Phone: %s, Avatar: %s", u.ID, u.OpenID, u.Nickname, u.Phone, u.Avatar)
}
