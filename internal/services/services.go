// Package services provides services for the application.
package services

import (
	"log/slog"

	"aibuddy/pkg/config"
	"aibuddy/pkg/flash"
)

// Flash is the global flash instance.
var Flash flash.Flash

// init initializes the flash instance.
func init() {
	var err error
	Flash, err = newFlash()
	if err != nil {
		slog.Error("failed to initialize flash", "error", err)
	}
}

// newFlash creates a new flash instance based on config.
func newFlash() (flash.Flash, error) {
	cfg := config.Instance.Storage.Flash
	redisCfg := config.Instance.Storage.Redis

	// Default to memory if no config
	if cfg == nil || cfg.Use == "memory" {
		return flash.NewMemory()
	}

	return flash.New(cfg.Use,
		flash.WithRedisConfig(redisCfg.Host, redisCfg.Port, redisCfg.Password, redisCfg.DB),
	)
}
