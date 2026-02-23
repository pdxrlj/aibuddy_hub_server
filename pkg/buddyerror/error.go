// Package buddyerror 提供错误类型定义
package buddyerror

import "errors"

// ErrorType 错误类型
type ErrorType error

// BuddyError 错误结构
type BuddyError struct {
	ErrorType ErrorType
	ErrorCode int
}

// BuddyErrorDict 错误字典
var BuddyErrorDict = map[ErrorType]BuddyError{
	ErrNoPermission: {
		ErrorType: ErrNoPermission,
		ErrorCode: 401,
	},
	ErrParamError: {
		ErrorType: ErrParamError,
		ErrorCode: 400,
	},
	ErrUserNotFound: {
		ErrorType: ErrUserNotFound,
		ErrorCode: 404,
	},
}

// GetBuddyErrorCode 获取错误码
func GetBuddyErrorCode(errorType ErrorType) int {
	return BuddyErrorDict[errorType].ErrorCode
}

// ErrNoPermission 没有权限
var ErrNoPermission ErrorType = errors.New("没有权限")

// ErrParamError 参数错误
var ErrParamError ErrorType = errors.New("参数错误")

// ErrUserNotFound 用户不存在
var ErrUserNotFound ErrorType = errors.New("用户不存在")

// ErrInvalidToken 无效的token
var ErrInvalidToken ErrorType = errors.New("无效的token")
