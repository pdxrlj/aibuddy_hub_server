package storage

import (
	"aibuddy/pkg/config"
	"aibuddy/pkg/helpers"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	cfg := config.Setup("../../config")
	helpers.PP(cfg)
	storage := NewStorage(
		cfg.Storage.OSS.AccessKeyID,
		cfg.Storage.OSS.AccessKeySecret,
		cfg.Storage.OSS.Region,
		cfg.Storage.OSS.Endpoint,
		cfg.Storage.OSS.Bucket,
	)
	assert.NotNil(t, storage)

	url, err := storage.PresignURL(context.Background(), "uploads/10/20250819_123053_9JQUrzSq.jpg", 15*time.Minute, "image/resize,m_lfit,w_100,h_200")
	assert.NoError(t, err)
	assert.NotEmpty(t, url)
	t.Log("url:", url)
}
