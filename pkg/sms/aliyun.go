// Package sms 阿里与短信
package sms

import (
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi20170525 "github.com/alibabacloud-go/dysmsapi-20170525/v5/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/credentials-go/credentials"
)

// AliyunSMS 配置参数
type AliyunSMS struct {
	AccessKeyID     string
	AccessKeySecret string
	SignName        string
	TemplateCode    string
	TemplateParam   string

	_Client *dysmsapi20170525.Client
}

// AliyunSMSOptions 短信配置
type AliyunSMSOptions func(*AliyunSMS)

// WithAccessKeyID 设置AccessKeyID
func WithAccessKeyID(accessKeyID string) AliyunSMSOptions {
	return func(sms *AliyunSMS) {
		sms.AccessKeyID = accessKeyID
	}
}

// WithAccessKeySecret 设置AccessKeySecret
func WithAccessKeySecret(accessKeySecret string) AliyunSMSOptions {
	return func(sms *AliyunSMS) {
		sms.AccessKeySecret = accessKeySecret
	}
}

// WithSignName 设置短信发送时的签名名称
func WithSignName(signName string) AliyunSMSOptions {
	return func(sms *AliyunSMS) {
		sms.SignName = signName
	}
}

// WithTemplateCode 设置短信发送时的模版code
func WithTemplateCode(templateCode string) AliyunSMSOptions {
	return func(sms *AliyunSMS) {
		sms.TemplateCode = templateCode
	}
}

// NewAliyunSMS 新的 AliyunSMS 实例
func NewAliyunSMS(options ...AliyunSMSOptions) (*AliyunSMS, error) {
	aliyun := &AliyunSMS{}

	for _, option := range options {
		option(aliyun)
	}

	CredentialsConfig := new(credentials.Config).
		SetType("access_key").
		SetAccessKeyId(aliyun.AccessKeyID).
		SetAccessKeySecret(aliyun.AccessKeySecret)

	credential, err := credentials.NewCredential(CredentialsConfig)
	if err != nil {
		return nil, err
	}

	config := &openapi.Config{
		Credential: credential,
	}
	config.Endpoint = tea.String("dysmsapi.aliyuncs.com")
	dysmsapiClient, err := dysmsapi20170525.NewClient(config)

	if err != nil {
		return nil, err
	}
	return &AliyunSMS{
		_Client:      dysmsapiClient,
		SignName:     aliyun.SignName,
		TemplateCode: aliyun.TemplateCode,
	}, err
}

// SendSMSResponse 发送短信接口的响应数据结构
type SendSMSResponse struct {
	Message   string `json:"Message"`
	RequestID string `json:"RequestId"`
	Code      string `json:"Code"`
	BizID     string `json:"BizId"`
}

// SendSMS 发送验证码短信
func (s *AliyunSMS) SendSMS(phoneNumbers string, code string) (*SendSMSResponse, error) {
	sendSmsRequest := &dysmsapi20170525.SendSmsRequest{
		SignName:      tea.String(s.SignName),
		TemplateCode:  tea.String(s.TemplateCode),
		TemplateParam: tea.String("{\"code\":\"" + code + "\"}"),
		PhoneNumbers:  tea.String(phoneNumbers),
	}

	runtime := &util.RuntimeOptions{}
	resp, err := s._Client.SendSmsWithOptions(sendSmsRequest, runtime)
	if err != nil {
		return nil, err
	}

	return &SendSMSResponse{
		Message:   tea.StringValue(resp.Body.Message),
		RequestID: tea.StringValue(resp.Body.RequestId),
		Code:      tea.StringValue(resp.Body.Code),
		BizID:     tea.StringValue(resp.Body.BizId),
	}, nil
}

// SendSMSWithDeviceMessage 发送消息至设备
func (s *AliyunSMS) SendSMSWithDeviceMessage(phoneNumbers string, childName string) (*SendSMSResponse, error) {
	sendSmsRequest := &dysmsapi20170525.SendSmsRequest{
		SignName:      tea.String(s.SignName),
		TemplateCode:  tea.String(s.TemplateCode),
		TemplateParam: tea.String("{\"child\":\"" + childName + "\"}"),
		PhoneNumbers:  tea.String(phoneNumbers),
	}

	runtime := &util.RuntimeOptions{}
	resp, err := s._Client.SendSmsWithOptions(sendSmsRequest, runtime)
	if err != nil {
		return nil, err
	}

	return &SendSMSResponse{
		Message:   tea.StringValue(resp.Body.Message),
		RequestID: tea.StringValue(resp.Body.RequestId),
		Code:      tea.StringValue(resp.Body.Code),
		BizID:     tea.StringValue(resp.Body.BizId),
	}, nil
}
