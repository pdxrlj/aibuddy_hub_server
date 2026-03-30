package wxmini

import (
	"aibuddy/pkg/config"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// 辅助函数：初始化 Template 实例
func getTestTemplate(t *testing.T) *Template {
	if config.Instance == nil {
		t.Fatal("config.Instance is not initialized")
	}

	wechatConfig := config.Instance.Wechat
	if wechatConfig == nil {
		t.Fatal("Wechat config is not loaded")
	}

	return NewTemplate(wechatConfig.AppID, wechatConfig.AppSecret)
}

// 辅助函数：检查是否为授权错误
func isAuthError(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "user refuse to accept") || strings.Contains(err.Error(), "43101"))
}

func Test_Reminie(t *testing.T) {
	template := getTestTemplate(t)
	err := template.SendMemoRemind("oNIUD7BmTr_Jtzofqutu4_rNVNpU", "这是一条测试消息", "remind", "67a0Y52uQoBLzep_Aqj-OzF-PTS98-WgkuGlT7_onPA")

	if isAuthError(err) {
		t.Skip("用户未授权订阅消息，跳过测试")
	}
	assert.NoError(t, err)
}

// 测试备忘录提醒
func Test_SendMemoRemind(t *testing.T) {
	template := getTestTemplate(t)
	err := template.SendMemoRemind(
		"oNIUD7BmTr_Jtzofqutu4_rNVNpU",
		"测试标题",
		"测试提醒内容",
		"67a0Y52uQoBLzep_Aqj-OzF-PTS98-WgkuGlT7_onPA",
	)

	if isAuthError(err) {
		t.Skip("用户未授权订阅消息，跳过测试")
	}
	assert.NoError(t, err)
}

// 测试留言提醒
func Test_SendMessageRemind(t *testing.T) {
	template := getTestTemplate(t)
	err := template.SendMessageRemind(
		"oNIUD7BmTr_Jtzofqutu4_rNVNpU",
		"张三",
		"这是一条留言内容",
		"yWL54YM--E9EpziRWZUtAcc1WsmEglbckDd4H-GaGCI",
	)

	if isAuthError(err) {
		t.Skip("用户未授权订阅消息，跳过测试")
	}
	assert.NoError(t, err)
}

// 测试留言通知
func Test_SendMessageNotify(t *testing.T) {
	template := getTestTemplate(t)
	err := template.SendMessageNotify(
		"oNIUD7BmTr_Jtzofqutu4_rNVNpU",
		"通知标题",
		"李四",
		"这是通知内容",
		"M4eG62rCInZ3gctzzWvn9832qLGBMDWHNzOKNzr-YjA",
	)

	if isAuthError(err) {
		t.Skip("用户未授权订阅消息，跳过测试")
	}
	assert.NoError(t, err)
}

// 测试生日提醒
func Test_BirthdayReminder(t *testing.T) {
	template := getTestTemplate(t)
	err := template.BirthdayReminder(
		"oNIUD7BmTr_Jtzofqutu4_rNVNpU",
		"用户A",
		"张小明",
		"生日备注信息",
		"记得买生日蛋糕",
		"0Vl6AGvLpnuzr1CfZUjDfDNLqa5yN2hcAd0YTxaRfKY",
	)

	if isAuthError(err) {
		t.Skip("用户未授权订阅消息，跳过测试")
	}
	assert.NoError(t, err)
}

// 测试会员到期提醒
func Test_MembershipExpirationReminder(t *testing.T) {
	template := getTestTemplate(t)
	expiration := time.Now().AddDate(0, 0, 7) // 7天后到期
	err := template.MembershipExpirationReminder(
		"oNIUD7BmTr_Jtzofqutu4_rNVNpU",
		"测试会员",
		expiration,
		"3PcDepqVpaImJZ4me2UShuk4TPVThPKYx6fH7qTtWS8",
	)

	if isAuthError(err) {
		t.Skip("用户未授权订阅消息，跳过测试")
	}
	assert.NoError(t, err)
}

// 测试开通会员成功通知
func Test_MembershipActivationSuccessfulNotification(t *testing.T) {
	template := getTestTemplate(t)
	startTime := time.Now()
	endTime := startTime.AddDate(0, 1, 0) // 一个月后
	err := template.MembershipActivationSuccessfulNotification(
		"oNIUD7BmTr_Jtzofqutu4_rNVNpU",
		"月度会员",
		"9.90",
		startTime,
		endTime,
		"QZf2PQ_05DwxzF1Jywx09hsFhNYZjEKpPuPa9bzh82A",
	)

	if isAuthError(err) {
		t.Skip("用户未授权订阅消息，跳过测试")
	}
	assert.NoError(t, err)
}

// 测试风险预警通知
func Test_RiskWarningNotice(t *testing.T) {
	template := getTestTemplate(t)
	err := template.RiskWarningNotice(
		"oNIUD7BmTr_Jtzofqutu4_rNVNpU",
		"高风险预警",
		"高风险",
		"检测到异常登录行为",
		time.Now().Format(time.DateTime),
		"p6FFFA3s-A7ktpHBakmifsceTHahKAoAXjIs4O5hH4Y",
	)

	if isAuthError(err) {
		t.Skip("用户未授权订阅消息，跳过测试")
	}
	assert.NoError(t, err)
}

// 批量测试所有模板消息
func Test_AllTemplateMessages(t *testing.T) {
	template := getTestTemplate(t)
	openID := "oNIUD7BmTr_Jtzofqutu4_rNVNpU"

	tests := []struct {
		name string
		fn   func() error
	}{
		{
			name: "备忘录提醒",
			fn: func() error {
				return template.SendMemoRemind(
					openID,
					"测试标题",
					"测试提醒",
					"67a0Y52uQoBLzep_Aqj-OzF-PTS98-WgkuGlT7_onPA",
				)
			},
		},
		{
			name: "留言提醒",
			fn: func() error {
				return template.SendMessageRemind(
					openID,
					"张三",
					"留言内容",
					"yWL54YM--E9EpziRWZUtAcc1WsmEglbckDd4H-GaGCI",
				)
			},
		},
		{
			name: "留言通知",
			fn: func() error {
				return template.SendMessageNotify(
					openID,
					"通知标题",
					"李四",
					"通知内容",
					"M4eG62rCInZ3gctzzWvn9832qLGBMDWHNzOKNzr-YjA",
				)
			},
		},
		{
			name: "生日提醒",
			fn: func() error {
				return template.BirthdayReminder(
					openID,
					"用户A",
					"张小明",
					"备注",
					"内容",
					"0Vl6AGvLpnuzr1CfZUjDfDNLqa5yN2hcAd0YTxaRfKY",
				)
			},
		},
		{
			name: "会员到期提醒",
			fn: func() error {
				return template.MembershipExpirationReminder(
					openID,
					"会员名称",
					time.Now().AddDate(0, 0, 7),
					"3PcDepqVpaImJZ4me2UShuk4TPVThPKYx6fH7qTtWS8",
				)
			},
		},
		{
			name: "开通会员成功通知",
			fn: func() error {
				startTime := time.Now()
				return template.MembershipActivationSuccessfulNotification(
					openID,
					"会员商品",
					"99.00",
					startTime,
					startTime.AddDate(1, 0, 0),
					"QZf2PQ_05DwxzF1Jywx09hsFhNYZjEKpPuPa9bzh82A",
				)
			},
		},
		{
			name: "风险预警通知",
			fn: func() error {
				return template.RiskWarningNotice(
					openID,
					"预警标题",
					"高风险",
					"预警内容",
					time.Now().Format(time.DateTime),
					"p6FFFA3s-A7ktpHBakmifsceTHahKAoAXjIs4O5hH4Y",
				)
			},
		},
	}

	authErrorCount := 0
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			if isAuthError(err) {
				authErrorCount++
				t.Skip("用户未授权订阅消息，跳过测试")
			}
			assert.NoError(t, err)
		})
	}

	// 如果所有测试都因为授权问题跳过，输出提示
	if authErrorCount == len(tests) {
		t.Log("所有测试都因为用户未授权而跳过，请确保：")
		t.Log("1. 在小程序端调用 wx.requestSubscribeMessage 获取用户授权")
		t.Log("2. 使用已授权的 openid 进行测试")
	}
}
