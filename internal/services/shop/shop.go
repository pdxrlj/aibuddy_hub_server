// Package shop 商城服务层
package shop

import (
	"aibuddy/internal/services/cache"
	"aibuddy/pkg/config"
	"aibuddy/pkg/flash"
	"aibuddy/pkg/shop"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// tracer 获取tracer
var tracer = func() trace.Tracer {
	return otel.Tracer(config.Instance.Tracer.ServiceName)
}

// Service 商城Service
type Service struct {
	cache    flash.Flash
	MiniShop *shop.MiniShop
}

// NewShopService 实例化
func NewShopService() *Service {
	return &Service{
		cache:    cache.Flash(),
		MiniShop: shop.NewMiniShop(),
	}
}

// GoodsList 获取商城列表
func (s *Service) GoodsList(ctx context.Context, pageSize int, nextKey string) ([]*shop.ProductProductResponse, int, string, error) {
	var goodsList []*shop.ProductProductResponse
	ctx, span := tracer().Start(ctx, "GoodsList")
	defer span.End()

	accessToken, err := s.GetShopAccessToken(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, 0, "", err
	}

	result, err := s.MiniShop.GetGoodsList(accessToken, pageSize, nextKey)
	if err != nil {
		span.RecordError(err)
		return nil, 0, "", err
	}

	for _, id := range result.ProductIDs {
		info, err := s.MiniShop.GetGoodsInfo(accessToken, id)
		if err != nil {
			span.RecordError(err)
			return nil, 0, "", err
		}
		goodsList = append(goodsList, info)
	}

	return goodsList, result.TotalNum, result.NextKey, nil
}

// GetShopAccessToken 获取商店token
func (s *Service) GetShopAccessToken(ctx context.Context) (string, error) {
	ctx, span := tracer().Start(ctx, "GetShopAccessToken")
	defer span.End()

	key := fmt.Sprintf("wx_shop_%s", config.Instance.MiniShop.AppID)
	token, err := s.cache.Get(key)
	if err != nil && !errors.Is(err, redis.Nil) {
		return "", err
	}
	if token != "" {
		return token.(string), nil
	}

	// 获取新的token
	data, err := shop.GetAccessToken(ctx, config.Instance.MiniShop.AppID, config.Instance.MiniShop.AppSecret)
	if err != nil {
		return "", err
	}

	if data.ErrorCode != 0 {
		return "", errors.New(data.ErrMsg)
	}

	token = data.AccessToken
	if err := s.cache.Set(key, token.(string), time.Duration(data.ExpiresIn)); err != nil {
		return "", err
	}

	return token.(string), nil
}
