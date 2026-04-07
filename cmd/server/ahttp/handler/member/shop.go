// Package membershop 会员商城
package membershop

import (
	"aibuddy/internal/services/aiuser"
	"aibuddy/internal/services/membershop"
	"aibuddy/pkg/ahttp"
	"aibuddy/pkg/config"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = func() trace.Tracer {
	return otel.Tracer(config.Instance.Tracer.ServiceName)
}

// Handler 商城handler
type Handler struct {
	MemberService *membershop.Service
}

// NewHandler 实例化handler
func NewHandler() *Handler {
	return &Handler{
		MemberService: membershop.NewService(),
	}
}

// GoodsList 获取商品列表
func (h *Handler) GoodsList(state *ahttp.State, req *GoodsListRequest) error {
	ctx, span := tracer().Start(state.Ctx.Request().Context(), "Shop.GoodsList")
	defer span.End()

	data, err := h.MemberService.GoodsList(ctx, req.Page, req.PageSize)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.Int("page", req.Page), attribute.Int("page_size", req.PageSize))
		return state.Response().Error(err)
	}
	return state.Response().Success(data)
}

// CreateOrder 创建订单
func (h *Handler) CreateOrder(state *ahttp.State, req *CreateOrderRequest) error {
	ctx, span := tracer().Start(state.Ctx.Request().Context(), "Shop.CreateOrder")
	defer span.End()

	uid, err := aiuser.GetUIDFromContext(state.Ctx)
	if err != nil {
		span.RecordError(err)
		return state.Response().SetStatus(http.StatusBadRequest).Error(err)
	}

	order, err := h.MemberService.CreateOrder(ctx, uid, req.GoodsID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.Int64("goods_id", req.GoodsID), attribute.Int64("uid", uid))
		return state.Response().Error(err)
	}
	return state.Response().Success(order)
}
