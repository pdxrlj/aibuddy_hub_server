package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetup(t *testing.T) {
	cfg := Setup("")
	assert.NotNil(t, cfg)
	assert.Equal(t, "aibuddy_hub", cfg.App.Name)
	assert.Equal(t, "0.0.0.0", cfg.App.Host)
	assert.Equal(t, "9081", cfg.App.Port)
	assert.Equal(t, "info", cfg.App.LogLevel)
}

func Test_WhenTestConfig_ThenConfigIsLoaded(t *testing.T) {
	cfg := Setup("")
	assert.NotNil(t, cfg)
	assert.Equal(t, "aibuddy_hub", cfg.App.Name)
	assert.Equal(t, "0.0.0.0", cfg.App.Host)
	assert.Equal(t, "9081", cfg.App.Port)
	assert.Equal(t, "info", cfg.App.LogLevel)
	assert.Equal(t, "8.153.82.116", cfg.Storage.Database.Host)
	assert.Equal(t, 25432, cfg.Storage.Database.Port)
	assert.Equal(t, "aibuddy", cfg.Storage.Database.Name)
	assert.Equal(t, "aibuddy", cfg.Storage.Database.User)
	assert.Equal(t, "aibuddy123456", cfg.Storage.Database.Password)
	assert.Equal(t, "127.0.0.1", cfg.Storage.Redis.Host)
	assert.Equal(t, 6379, cfg.Storage.Redis.Port)
	assert.Equal(t, "test", cfg.Storage.Redis.Password)
	assert.Equal(t, "redis", cfg.Storage.Flash.Use)

	assert.Equal(t, "doubao-seed-1-8-251228", cfg.Agent.Model.ChatModel.ModelName)
	assert.Equal(t, "doubao-seed-1-8-251228", cfg.Agent.Model.WorkModel.ModelName)
	assert.Equal(t, "xx", cfg.Agent.Model.ChatModel.APIKey)
	assert.Equal(t, "xx", cfg.Agent.Model.WorkModel.APIURL)
}
