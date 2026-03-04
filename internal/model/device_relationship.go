// Package model 用户关系模型
package model

import "time"

// RelationshipStatus 关系状态
type RelationshipStatus string

const (
	// RelationshipStatusPending 等待确认
	RelationshipStatusPending RelationshipStatus = "等待确认" // 等待确认
	// RelationshipStatusAccepted 已接受
	RelationshipStatusAccepted RelationshipStatus = "已接受" // 已接受
	// RelationshipStatusRejected 已拒绝
	RelationshipStatusRejected RelationshipStatus = "已拒绝" // 已拒绝
	// RelationshipStatusBlocked 已拉黑
	RelationshipStatusBlocked RelationshipStatus = "已拉黑" // 已拉黑
)

// DeviceRelationship 设备关系表（好友关系）
type DeviceRelationship struct {
	ID int64 `gorm:"primaryKey;autoIncrement;column:id;"`

	// 发起方用户ID
	DeviceID string `gorm:"column:device_id;index;type:varchar(255);not null;comment:发起方用户ID;"`
	// 目标用户ID
	TargetDeviceID string `gorm:"column:target_device_id;index;type:varchar(255);not null;comment:目标用户ID;"`

	// 关系状态
	Status RelationshipStatus `gorm:"column:status;type:varchar(16);not null;default:pending;comment:关系状态;"`

	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;not null;comment:创建时间;"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;not null;comment:更新时间;"`
}

// TableName 表名
func (DeviceRelationship) TableName() string {
	return TableName("device_relationship")
}

// IsFriend 是否是好友关系
func (r *DeviceRelationship) IsFriend() bool {
	return r.Status == RelationshipStatusAccepted
}

// IsPending 是否等待中
func (r *DeviceRelationship) IsPending() bool {
	return r.Status == RelationshipStatusPending
}

// CanAccept 是否可以接受（只有目标用户可以接受）
func (r *DeviceRelationship) CanAccept(targetDeviceID string) bool {
	return r.Status == RelationshipStatusPending && r.TargetDeviceID == targetDeviceID
}

// CanReject 是否可以拒绝（只有目标用户可以拒绝）
func (r *DeviceRelationship) CanReject(targetDeviceID string) bool {
	return r.Status == RelationshipStatusPending && r.TargetDeviceID == targetDeviceID
}

// Accept 接受好友请求
func (r *DeviceRelationship) Accept() {
	now := time.Now()
	r.Status = RelationshipStatusAccepted
	r.UpdatedAt = now
}

// Reject 拒绝好友请求
func (r *DeviceRelationship) Reject() {
	now := time.Now()
	r.Status = RelationshipStatusRejected
	r.UpdatedAt = now
}
