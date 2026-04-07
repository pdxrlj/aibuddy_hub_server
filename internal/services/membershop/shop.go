// Package membershop 提供会员商城相关的业务逻辑
package membershop

import (
	"aibuddy/internal/model"
	"aibuddy/internal/repository"
	"aibuddy/pkg/config"
	"aibuddy/pkg/pay"
	"aibuddy/pkg/pay/certs"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/go-pay/gopay/wechat/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

// DefaultFirstFreeVipTime 首次激活默认赠送一年VIP
const DefaultFirstFreeVipTime = 365 * 24 * 60 * 60

// tracer 获取tracer
var tracer = func() trace.Tracer {
	return otel.Tracer(config.Instance.Tracer.ServiceName)
}

// Service 商城服务层
type Service struct {
	MemberShopRepository *repository.MemberShopRepository
	UserRepository       *repository.UserRepo
	WxMinPayService      *pay.WxMinPay
}

// NewService 实例化
func NewService() *Service {
	cfg := config.Instance.Pay

	WxMinPayService, err := pay.NewWxMinPay(
		pay.WithAppID(cfg.AppID),
		pay.WithMchID(cfg.MchID),
		pay.WithAPIV3Key(cfg.APIV3Key),
		pay.WithPrivateKey(certs.ApiclientKey),
		pay.WithSerialNo(cfg.SerialNo),
		pay.WithWxPublicKey(cfg.WechatpaySerialNo),
		pay.WithWxPublicKeyContent([]byte(certs.WechatpayPublicKey)),
		pay.WithDebug(cfg.Debug),
		pay.WithOrderNotifyURL(cfg.NotifyURL),
		pay.WithRefundNotifyURL(cfg.RefundNotifyURL),
	)

	if err != nil {
		panic(err)
	}

	return &Service{
		MemberShopRepository: repository.NewMemberShopRepository(),
		UserRepository:       repository.NewUserRepo(),
		WxMinPayService:      WxMinPayService,
	}
}

// GoodsListResponse 商品列表响应
type GoodsListResponse struct {
	Total int64
	List  []*model.Goods
}

// GoodsList 获取商品列表
func (s *Service) GoodsList(ctx context.Context, page, pageSize int) (*GoodsListResponse, error) {
	_, span := tracer().Start(ctx, "ShopService.GoodsList")
	defer span.End()

	goods, total, err := s.MemberShopRepository.GoodsList(ctx, page, pageSize)
	return &GoodsListResponse{
		Total: total,
		List:  goods,
	}, err
}

// CreateOrder 创建订单
func (s *Service) CreateOrder(ctx context.Context, userID, goodsID int64) (*CreateOrderResponse, error) {
	slog.Info("[Shop] CreateOrder", "userID", userID, "goodsID", goodsID)
	ctx, span := tracer().Start(ctx, "ShopService.CreateOrder")
	defer span.End()

	// 1. 验证商品
	goods, err := s.ValidateGoods(ctx, goodsID)
	if err != nil {
		return nil, err
	}

	// 2. 获取用户信息并验证
	userInfo, err := s.GetUserInfo(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 3. 计算价格并生成订单号
	price := s.CalculatePrice(goods)
	orderNo := s.GenerateOrderNo(userID)

	// 4. 执行事务创建订单
	wxPreResp, err := s.ExecuteOrderTransaction(ctx, userID, goods, orderNo, price, userInfo.OpenID)
	if err != nil {
		return nil, err
	}

	// 5. 生成支付签名并返回
	return s.BuildOrderResponse(orderNo, wxPreResp, price, goods.Name)
}

// GetUserInfo 获取用户信息并验证
func (s *Service) GetUserInfo(ctx context.Context, userID int64) (*model.User, error) {
	_, span := tracer().Start(ctx, "ShopService.getUserInfo")
	defer span.End()

	userInfo, err := s.UserRepository.FindUserByUserID(userID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.Int64("userID", userID))
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	if userInfo == nil || userInfo.OpenID == "" {
		err := errors.New("用户未绑定微信，无法支付")
		span.RecordError(err)
		span.SetAttributes(attribute.Int64("userID", userID))
		return nil, err
	}

	return userInfo, nil
}

// CalculatePrice 计算价格（支持活动价）
func (s *Service) CalculatePrice(goods *model.Goods) int64 {
	if goods.ActivityPrice > 0 {
		return goods.ActivityPrice
	}
	return goods.Price
}

// ExecuteOrderTransaction 执行订单事务
func (s *Service) ExecuteOrderTransaction(ctx context.Context, userID int64, goods *model.Goods, orderNo string, price int64, openID string) (*wechat.PrepayRsp, error) {
	_, span := tracer().Start(ctx, "ShopService.ExecuteOrderTransaction")
	defer span.End()

	var wxPreResp *wechat.PrepayRsp

	err := s.MemberShopRepository.Transaction(func(tx *gorm.DB) error {
		// 创建订单记录
		order, err := s.CreateOrderRecord(tx, userID, goods, orderNo, price)
		if err != nil {
			return fmt.Errorf("创建订单记录失败: %w", err)
		}

		// 扣减库存
		if err := s.DeductStock(tx, goods.ID); err != nil {
			return fmt.Errorf("扣减库存失败: %w", err)
		}

		// 创建微信预支付订单
		wxResp, err := s.CreateWxPrepayOrder(ctx, orderNo, goods.Name, price, openID)
		if err != nil {
			return fmt.Errorf("创建微信预支付订单失败: %w", err)
		}

		wxPreResp = wxResp

		// 更新订单的微信订单号
		if err := s.UpdateOutTradeNo(tx, order.ID, wxPreResp.Response.PrepayId); err != nil {
			return fmt.Errorf("更新订单号失败: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return wxPreResp, nil
}

// BuildOrderResponse 构建订单响应
func (s *Service) BuildOrderResponse(orderNo string, wxPreResp *wechat.PrepayRsp, price int64, goodsName string) (*CreateOrderResponse, error) {
	paySign, err := s.WxMinPayService.PaySignOfJSAPI(s.WxMinPayService.AppID, wxPreResp.Response.PrepayId)
	if err != nil {
		return nil, fmt.Errorf("生成支付签名失败: %w", err)
	}

	return &CreateOrderResponse{
		OrderNo:   orderNo,
		PrepayID:  wxPreResp.Response.PrepayId,
		PaySign:   paySign,
		Amount:    price,
		GoodsName: goodsName,
	}, nil
}

// CreateOrderResponse 创建订单响应
type CreateOrderResponse struct {
	OrderNo   string `json:"order_no"`   // 商户订单号
	PrepayID  string `json:"prepay_id"`  // 微信预支付ID
	PaySign   any    `json:"pay_sign"`   // 支付签名
	Amount    int64  `json:"amount"`     // 支付金额（分）
	GoodsName string `json:"goods_name"` // 商品名称
}

// ValidateGoods 验证商品
func (s *Service) ValidateGoods(ctx context.Context, goodsID int64) (*model.Goods, error) {
	goods, err := s.MemberShopRepository.GoodsInfo(ctx, goodsID)
	if err != nil {
		return nil, fmt.Errorf("获取商品信息失败: %w", err)
	}

	if goods.Status != model.GoodsStatusNormal {
		return nil, errors.New("商品已下架")
	}

	if goods.Stock <= 0 {
		return nil, errors.New("商品库存不足")
	}

	return goods, nil
}

// GenerateOrderNo 生成订单号（格式：ORD + 时间戳 + 用户ID后4位 + 4位随机数）
func (s *Service) GenerateOrderNo(userID int64) string {
	userIDSuffix := userID % 10000
	randomNum := rand.Intn(10000)
	return fmt.Sprintf("ORD%s%04d%04d", time.Now().Format("20060102150405"), userIDSuffix, randomNum)
}

// CreateOrderRecord 创建订单记录
func (s *Service) CreateOrderRecord(tx *gorm.DB, userID int64, goods *model.Goods, orderNo string, price int64) (*model.Order, error) {
	order := &model.Order{
		UserID:  userID,
		OrderNo: orderNo,
		Status:  model.OrderStatusPending,
		Goods: []*model.OrderGoods{
			{
				GoodsID:    goods.ID,
				GoodsName:  goods.Name,
				GoodsPrice: price,
				GoodsNum:   1,
			},
		},
	}

	if err := tx.Create(order).Error; err != nil {
		return nil, err
	}

	return order, nil
}

// DeductStock 扣减库存
func (s *Service) DeductStock(tx *gorm.DB, goodsID int64) error {
	result := tx.Model(&model.Goods{}).
		Where("id = ? AND stock > 0", goodsID).
		Update("stock", gorm.Expr("stock - 1"))

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("库存不足")
	}

	return nil
}

// CreateWxPrepayOrder 创建微信预支付订单
func (s *Service) CreateWxPrepayOrder(ctx context.Context, orderNo, description string, amount int64, openID string) (*wechat.PrepayRsp, error) {
	orderReq := pay.DefaultWxOrderRequest()
	orderReq.Description = description
	orderReq.OutTradeNo = orderNo
	orderReq.TotalAmount = amount
	orderReq.Payer.OpenID = openID

	orderInfo, err := s.WxMinPayService.CreateOrder(ctx, orderReq)
	if err != nil {
		return nil, err
	}

	wxPreResp, ok := orderInfo.(*wechat.PrepayRsp)
	if !ok {
		return nil, errors.New("微信预支付订单返回类型错误")
	}

	return wxPreResp, nil
}

// UpdateOutTradeNo 更新订单的微信订单号
func (s *Service) UpdateOutTradeNo(tx *gorm.DB, orderID int64, prepayID string) error {
	return tx.Model(&model.Order{}).
		Where("id = ?", orderID).
		Update("out_trade_no", prepayID).Error
}
