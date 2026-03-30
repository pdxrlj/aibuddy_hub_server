// Package wxmini 提供微信小程序相关功能
package wxmini

import (
	"strings"
	"time"

	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	"github.com/silenceper/wechat/v2/miniprogram"
	"github.com/silenceper/wechat/v2/miniprogram/config"
	"github.com/silenceper/wechat/v2/miniprogram/subscribe"
)

// Template 微信小程序模板消息服务
type Template struct {
	miniProgram *miniprogram.MiniProgram
}

// NewTemplate 创建微信小程序模板消息服务实例
func NewTemplate(appID, appSecret string) *Template {
	wc := wechat.NewWechat()
	memory := cache.NewMemory()
	miniProgram := wc.GetMiniProgram(&config.Config{
		Cache:     memory,
		AppID:     appID,
		AppSecret: appSecret,
	})
	return &Template{
		miniProgram: miniProgram,
	}
}

// SendMemoRemind 发送备忘录提醒通知
func (t *Template) SendMemoRemind(openID string, title, remind string, templateID string) error {
	return t.miniProgram.GetSubscribe().Send(
		&subscribe.Message{
			ToUser:     openID,
			TemplateID: templateID,
			Data: map[string]*subscribe.DataItem{
				"thing5": {
					Value: title,
				},
				"thing3": {
					Value: remind,
				},
				"time8": {
					Value: time.Now().Format(time.DateTime),
				},
			},
		},
	)
}

// SendMessageRemind 发送留言提醒通知
func (t *Template) SendMessageRemind(openID string, fromUser, message string, templateID string) error {
	return t.miniProgram.GetSubscribe().Send(
		&subscribe.Message{
			ToUser:     openID,
			TemplateID: templateID,
			Data: map[string]*subscribe.DataItem{
				"name1": {
					Value: fromUser,
				},
				"thing2": {
					Value: message,
				},
				"date3": {
					Value: time.Now().Format(time.DateTime),
				},
			},
		},
	)
}

// SendMessageNotify 发送留言通知
func (t *Template) SendMessageNotify(openID string, notify string, fromUser, message string, templateID string) error {
	return t.miniProgram.GetSubscribe().Send(
		&subscribe.Message{
			ToUser:     openID,
			TemplateID: templateID,
			Data: map[string]*subscribe.DataItem{
				"thing14": {
					Value: notify,
				},
				"name1": {
					Value: fromUser,
				},
				"thing2": {
					Value: message,
				},
				"time3": {
					Value: time.Now().Format(time.DateTime),
				},
			},
		},
	)
}

// BirthdayReminder 发送生日提醒通知
func (t *Template) BirthdayReminder(openID string, username, _ string, note, content, templateID string) error {
	return t.miniProgram.GetSubscribe().Send(
		&subscribe.Message{
			ToUser:     openID,
			TemplateID: templateID,
			Data: map[string]*subscribe.DataItem{
				"name1": {
					Value: username,
				},
				"time5": {
					Value: time.Now().Format(time.DateTime),
				},
				"thing3": {
					Value: note,
				},
				"thing6": {
					Value: content,
				},
			},
		},
	)
}

// MembershipExpirationReminder 发送会员到期提醒通知
func (t *Template) MembershipExpirationReminder(openID string, memberName string, expiration time.Time, templateID string) error {
	return t.miniProgram.GetSubscribe().Send(
		&subscribe.Message{
			ToUser:     openID,
			TemplateID: templateID,
			Data: map[string]*subscribe.DataItem{
				"thing1": {
					Value: memberName,
				},
				"date2": {
					Value: expiration.Format(time.DateTime),
				},
			},
		},
	)
}

// MembershipActivationSuccessfulNotification 发送开通会员成功通知
func (t *Template) MembershipActivationSuccessfulNotification(openID string, goodsName, amount string, startTime, endTime time.Time, templateID string) error {
	return t.miniProgram.GetSubscribe().Send(
		&subscribe.Message{
			ToUser:     openID,
			TemplateID: templateID,
			Data: map[string]*subscribe.DataItem{
				"thing7": {
					Value: goodsName,
				},
				"amount2": {
					Value: amount,
				},
				"time5": {
					Value: startTime.Format(time.DateTime),
				},
				"date3": {
					Value: endTime.Format(time.DateTime),
				},
			},
		},
	)
}

// RiskWarningNotice 发送风险预警通知
func (t *Template) RiskWarningNotice(openID string, warningTitle, riskLevel, warningContent, warningTime string, templateID string) error {
	return t.miniProgram.GetSubscribe().Send(
		&subscribe.Message{
			ToUser:     openID,
			TemplateID: templateID,
			Data: map[string]*subscribe.DataItem{
				"phrase7": {
					Value: warningTitle,
				},
				"thing10": {
					Value: riskLevel,
				},
				"thing9": {
					Value: warningContent,
				},
				"time2": {
					Value: warningTime,
				},
			},
		},
	)
}

// IsAuthError 判断是否为用户拒绝授权错误
func IsAuthError(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "user refuse to accept") || strings.Contains(err.Error(), "43101"))
}
