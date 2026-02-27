// Package cache 提供缓存服务
package cache

import (
	"aibuddy/pkg/config"
	"aibuddy/pkg/flash"
	"sync"
)

var once sync.Once

var flashInstance flash.Flash

// Flash 返回全局Flash实例
var Flash = func() flash.Flash {
	once.Do(func() {
		var err error
		flashInstance, err = NewFlash()
		if err != nil {
			panic(err)
		}
	})
	return flashInstance
}

// NewFlash 创建Flash实例
func NewFlash() (flash.Flash, error) {
	cfg := config.Instance.Storage.Flash
	redisCfg := config.Instance.Storage.Redis

	// Default to memory if no config
	if cfg == nil || cfg.Use == "memory" {
		return flash.NewMemory()
	}

	return flash.New(cfg.Use,
		flash.WithRedisConfig(redisCfg.Host, redisCfg.Port, redisCfg.Username, redisCfg.Password, redisCfg.DB),
	)
}
