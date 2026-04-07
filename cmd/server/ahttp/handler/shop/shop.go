// Package shop handler层
package shop

import (
	"aibuddy/internal/services/shop"
	"aibuddy/pkg/ahttp"
	"aibuddy/pkg/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var tracer = func() trace.Tracer {
	return otel.Tracer(config.Instance.Tracer.ServiceName)
}

// Handler handler
type Handler struct {
	ShopService *shop.Service
}

// NewShopHandler 实例化handler
func NewShopHandler() *Handler {
	return &Handler{
		ShopService: shop.NewShopService(),
	}
}

// GoodsList 获取商品列表
func (s *Handler) GoodsList(state *ahttp.State, req *GoodsListRequest) error {
	ctx, span := tracer().Start(state.Ctx.Request().Context(), "Shop.GoodsList")
	defer span.End()

	data, total, nexKey, err := s.ShopService.GoodsList(ctx, req.PageSize, req.NextKey)
	if err != nil {
		return state.Response().Error(err)
	}

	return state.Response().SetData(&GoodsListResponse{
		Total:   total,
		NextKey: nexKey,
		List:    data,
	}).Success()
}
