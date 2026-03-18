// Package userhandler 提供用户相关的 HTTP 处理器
package userhandler

import (
	"fmt"
	"time"
)

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// NewLoginRequest 微信登录请求
type NewLoginRequest struct {
	Source        string `json:"source" validate:"required" msg:"required:选择登录方式"`  // 登录来源（mini=小程序,phone=手机号）
	WechatCode    string `json:"wechat_code" validate:"required_if=Source mini"`    // 微信登录临时 code
	EncryptedData string `json:"encrypted_data" validate:"required_if=Source mini"` // 微信加密数据
	IV            string `json:"iv" validate:"required_if=Source mini"`             // 微信加密数据的初始向量

	Phone     string `json:"phone" validate:"required_if=Source phone"`      // 手机号（仅 source=phone 时必填）
	PhoneCode string `json:"phone_code" validate:"required_if=Source phone"` // 手机验证码（仅 source=phone 时必填）
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token    string `json:"token"`
	Expires  int64  `json:"expires"`
	UID      int64  `json:"uid"`
	OpenID   string `json:"open_id,omitempty"`
	Nickname string `json:"nickname,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Avatar   string `json:"avatar"`
}

// SendCodeRequest 验证码请求
type SendCodeRequest struct {
	Phone string `json:"phone" validate:"required,chmobile" msg:"required:用户手机号码不能为空|chmobile:手机号格式无效"`
}

// TokenRequest 退出登录请求
type TokenRequest struct {
	Token string `json:"token" validate:"required"`
}

// TokenResponse Token请求响应
type TokenResponse struct {
	Token   string `json:"token"`
	Expires int64  `json:"expires"`
}

// UserinfoRequest 用户信息请求数据
type UserinfoRequest struct {
	ID        int64  `json:"id"`
	BoardType string `json:"board_type" validate:"required" msg:"required:板子类型不能为空"`
	Version   string `json:"version" validate:"required" msg:"required:版本号不能为空"`

	DeviceID string `json:"device_id" validate:"required"`
	NickName string `json:"nickname" validate:"required"`
	Avatar   string `json:"avatar"`
	Gender   string `json:"gender" validate:"required,oneof=未知 男 女"`
	Birthday string `json:"birthday" validate:"required"`
	Relation string `json:"relation"`

	Hobbies     string `json:"hobbies"`
	Values      string `json:"values"`
	Skills      string `json:"skills"`
	Personality string `json:"personality"`
}

// LostRequest 挂失请求
type LostRequest struct {
	DeviceID string `json:"device_id" form:"device_id" param:"device_id" query:"device_id" validate:"required,aimac" msg:"required:设备ID不能为空|aimac:设备ID格式无效"`
}

// UnlostRequest 解除挂失请求
type UnlostRequest struct {
	DeviceID string `json:"device_id" form:"device_id" param:"device_id" query:"device_id" validate:"required,aimac" msg:"required:设备ID不能为空|aimac:设备ID格式无效"`
}

// UnbindRequest 解绑请求
type UnbindRequest struct {
	DeviceID string `json:"device_id" form:"device_id" param:"device_id" query:"device_id" validate:"required,aimac" msg:"required:设备ID不能为空|aimac:设备ID格式无效"`
}

// HaveDeviceResponse 是否第一次完善设备信息响应
type HaveDeviceResponse struct {
	HaveDevice bool `json:"have_device"`
}

// DeviceListResponse 设备列表响应
type DeviceListResponse struct {
	DeviceList []*DeviceInfoListItem `json:"device_list"`
}

// DeviceInfoListItem 设备信息列表项
type DeviceInfoListItem struct {
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
	Version    string `json:"version"`
	Status     string `json:"status"`
	Avatar     string `json:"avatar"`
	Gender     string `json:"gender"`
}

// SendMsgRequest 创建留言数据
type SendMsgRequest struct {
	DeviceID string `json:"device_id" validate:"required,aimac" msg:"required:设备ID不能为空|aimac:设备ID格式无效"`
	Fmt      string `json:"fmt" validate:"required" msg:"required:信息格式不能为空"`
	Content  string `json:"content" validate:"required,max=50" msg:"required:信息内容不能为空|max:消息内容最长为50个字"`
	Dur      int    `json:"dur" validate:"required_if_gt=Fmt=voice" msg:"required_if_gt:语音消息时长必须大于0"`
}

// GetMessageRequest 获取留言列表数据
type GetMessageRequest struct {
	DeviceID string `json:"device_id" form:"device_id" param:"device_id" query:"device_id" validate:"required,aimac" msg:"required:设备ID不能为空|aimac:设备ID格式无效"`
	Page     int    `json:"page" form:"page" param:"page" query:"page" validate:"required"`
	PageSize int    `json:"page_size" form:"page_size" param:"page_size" query:"page_size" validate:"required"`
}

// MsgListResponse 留言列表响应
type MsgListResponse struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
	Data     any   `json:"data"`
}

// AnalysisGrowthReportRequest 分析用户成长报告请求
type AnalysisGrowthReportRequest struct {
	DeviceID string `json:"device_id" form:"device_id" param:"device_id" query:"device_id" validate:"required,aimac" msg:"required:设备ID不能为空|aimac:设备ID格式无效"`

	StartTime string `json:"start_time" form:"start_time" param:"start_time" query:"start_time" validate:"required" msg:"required:开始时间不能为空"`
	EndTime   string `json:"end_time" form:"end_time" param:"end_time" query:"end_time" validate:"required" msg:"required:结束时间不能为空"`
}

// ParseTime 解析时间字符串
func (r *AnalysisGrowthReportRequest) ParseTime() (startTime, endTime time.Time, err error) {
	layout := "2006-01-02 15:04:05"
	startTime, err = time.Parse(layout, r.StartTime)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("开始时间格式错误: %w", err)
	}
	endTime, err = time.Parse(layout, r.EndTime)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("结束时间格式错误: %w", err)
	}
	return startTime, endTime, nil
}

// GetGrowthReportListRequest 获取用户成长报告列表请求
type GetGrowthReportListRequest struct {
	DeviceID string `json:"device_id" form:"device_id" param:"device_id" query:"device_id" validate:"required,aimac" msg:"required:设备ID不能为空|aimac:设备ID格式无效"`
	Page     int    `json:"page" form:"page" param:"page" query:"page" validate:"required,min=1" msg:"required:页码不能为空|min:页码不能小于1"`
	PageSize int    `json:"page_size" form:"page_size" param:"page_size" query:"page_size" validate:"required,min=1" msg:"required:每页条数不能为空|min:每页条数不能小于1"`
}

// GrowthReportListResponse 成长报告列表响应
type GrowthReportListResponse struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
	Data     any   `json:"data"`
}

// InfoRequest 用户信息请求
type InfoRequest struct {
	Username string `json:"username" validate:"omitempty,min=2,max=6" msg:"min:姓名长度不能小于2|max:姓名最长长度为6"`
	Phone    string `json:"phone" validate:"omitempty"`
	Gender   *int   `json:"gender" validate:"omitempty,min=0,max=2" msg:"min:参数取值范围[0,2]|max:参数取值范围[0,2]"`
	Birthday string `json:"birthday" validate:"omitempty"`
	Email    string `json:"email" validate:"omitempty,email" msg:"email:邮箱格式不正确"`
	NickName string `json:"nickname" validate:"omitempty,required,min=2,max=8" msg:"required:昵称不能为空|min:昵称长度不能小于2|max:昵称最长长度为8"`
	Avatar   string `json:"avatar" validate:"omitempty"`
}

// InfoResponse 用户信息响应
type InfoResponse struct {
	UID      int64  `json:"uid"`
	Useranem string `json:"username"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Gender   int    `json:"gender"`
	Birthday string `json:"birthday"`
	NickName string `json:"nickname"`
	Avatar   string `json:"avatar"`
}
