// Package wechatservice 微信服务--小程序
package wechatservice

import (
	"fmt"
	"sync"

	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	"github.com/silenceper/wechat/v2/miniprogram"
	wechatConfig "github.com/silenceper/wechat/v2/miniprogram/config"
)

var (
	wechatMiniProgram *miniprogram.MiniProgram
	once              sync.Once
	initialized       bool
)

// InitWechatMiniProgram 初始化小程序
func InitWechatMiniProgram(appID, appSecret string) error {
	var initErr error
	// 验证 appId 和 appSecret 的有效性
	if appID == "" || appSecret == "" {
		initErr = fmt.Errorf("appId or appSecret is empty")
	}

	once.Do(func() {
		mini := wechat.NewWechat()
		memory := cache.NewMemory()
		wechatMiniProgram = mini.GetMiniProgram(&wechatConfig.Config{
			AppID:     appID,
			AppSecret: appSecret,
			Cache:     memory,
		})

		initialized = true
	})

	return initErr
}

// GetWechatMiniProgram 获取微信小程序实例
func GetWechatMiniProgram() (*miniprogram.MiniProgram, error) {
	if !initialized {
		return nil, fmt.Errorf("wechat mini program not initialized")
	}
	return wechatMiniProgram, nil
}
