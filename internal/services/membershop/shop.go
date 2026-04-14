// Package membershop 提供会员商城相关的业务逻辑
package membershop

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"aibuddy/internal/repository"
	"aibuddy/pkg/config"
	"aibuddy/pkg/pay"
	"aibuddy/pkg/pay/certs"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"strings"
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
	DeviceRepo           *repository.DeviceRepo
	OrderRepo            *repository.OrderRepo

	WxMinPayService *pay.WxMinPay
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
		DeviceRepo:           repository.NewDeviceRepo(),
		OrderRepo:            repository.NewOrderRepo(),

		WxMinPayService: WxMinPayService,
	}
}

// GoodsListResponse 商品列表响应
type GoodsListResponse struct {
	Total int64
	List  []*model.GoodsActivity
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
func (s *Service) CreateOrder(ctx context.Context, userID, goodsID int64, deviceID string) (*CreateOrderResponse, error) {
	slog.Info("[Shop] CreateOrder", "userID", userID, "goodsID", goodsID, "deviceID", deviceID)
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
	price, aid := s.CalculatePrice(ctx, userID, deviceID, goods)
	orderNo := s.GenerateOrderNo(userID)

	// 4. 执行事务创建订单
	wxPreResp, err := s.ExecuteOrderTransaction(ctx, userID, deviceID, goods, orderNo, price, userInfo.OpenID, aid)
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
func (s *Service) CalculatePrice(ctx context.Context, userID int64, deviceID string, goods *model.Goods) (int64, int64) {
	// 是否已开通会员
	ctx, span := tracer().Start(ctx, "ShopService.CalculatePrice")
	defer span.End()
	activityList, err := s.MemberShopRepository.GetActvityByGoodsID(ctx, goods.ID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.Int64("goodsID", goods.ID), attribute.String("msg", "暂无活动"))
		return goods.Price, 0
	}

	memberLevel := s.OrderRepo.GetMemberLevel(ctx, userID, deviceID)
	var aid int64
	for _, v := range activityList {
		switch v.Type {
		case 0:
			if v.Count > 0 {
				if err := s.MemberShopRepository.SubActivityCount(ctx, v.ID); err != nil {
					span.RecordError(err)
					return goods.Price, 0
				}
				goods.ActivityPrice = v.Pirce
				aid = v.ID
			}
		case 1:
			if memberLevel == 1 {
				goods.ActivityPrice = v.Pirce
				aid = v.ID
			}
		case 2:
			if memberLevel == 2 {
				goods.ActivityPrice = v.Pirce
				aid = v.ID
			}
		}
	}
	span.SetAttributes(attribute.Int64("goodsID", goods.ID),
		attribute.Int("memberLevel", memberLevel),
		attribute.Int64("activity_id", aid),
		attribute.Int64("order_price", goods.ActivityPrice),
	)
	return goods.ActivityPrice, aid
}

// ExecuteOrderTransaction 执行订单事务
func (s *Service) ExecuteOrderTransaction(ctx context.Context, userID int64, deviceID string, goods *model.Goods, orderNo string, price int64, openID string, aid int64) (*wechat.PrepayRsp, error) {
	_, span := tracer().Start(ctx, "ShopService.ExecuteOrderTransaction")
	defer span.End()

	var wxPreResp *wechat.PrepayRsp

	err := s.MemberShopRepository.Transaction(func(tx *gorm.DB) error {
		// 创建订单记录
		if _, err := s.CreateOrderRecord(tx, userID, deviceID, goods, orderNo, price, aid); err != nil {
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
func (s *Service) CreateOrderRecord(tx *gorm.DB, userID int64, deviceID string, goods *model.Goods, outTradeNo string, price int64, aid int64) (*model.Order, error) {
	order := &model.Order{
		UserID:     userID,
		DeviceID:   deviceID,
		OutTradeNo: outTradeNo,
		Status:     model.OrderStatusPending,
		ActivityID: aid,
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

// OrderList 订单列表响应
type OrderList struct {
	Total int              `json:"total"`
	List  []*OrderListItem `json:"list"`
}

// OrderListItem 订单列表项
type OrderListItem struct {
	OrderNo   string `json:"orderNo"` // 商户订单号
	Status    string `json:"status"`
	Amount    int64  `json:"amount"`
	CreatedAt string `json:"createdAt"`

	GoodsName  string `json:"goodsName"`
	GoodsNum   int64  `json:"goodsNum"`
	GoodsPrice int64  `json:"goodsPrice"`
}

// OrderList 获取订单列表
func (s *Service) OrderList(ctx context.Context, page, pageSize int, status string) (*OrderList, error) {
	ctx, span := tracer().Start(ctx, "ShopService.OrderList")
	defer span.End()

	orders, count, err := s.MemberShopRepository.OrderList(ctx, page, pageSize, status)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(
			attribute.Int("page", page),
			attribute.Int("pageSize", pageSize),
			attribute.String("status", status),
		)
		return nil, err
	}

	// 转换为简化的订单列表项
	list := make([]*OrderListItem, 0, len(orders))
	for _, order := range orders {
		item := &OrderListItem{
			OrderNo:   order.OutTradeNo, // 使用 OutTradeNo 作为商户订单号
			Status:    string(order.Status),
			CreatedAt: order.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		// 处理商品信息
		var totalAmount int64
		var totalNum int64
		var goodsNames []string
		for _, goods := range order.Goods {
			totalAmount += goods.GoodsPrice * goods.GoodsNum
			totalNum += goods.GoodsNum
			goodsNames = append(goodsNames, goods.GoodsName)
		}

		item.Amount = totalAmount
		item.GoodsNum = totalNum
		item.GoodsName = strings.Join(goodsNames, ",")
		// 如果只有一个商品，显示单价；多个商品显示总价
		if len(order.Goods) == 1 {
			item.GoodsPrice = order.Goods[0].GoodsPrice
		} else {
			item.GoodsPrice = totalAmount
		}

		list = append(list, item)
	}

	return &OrderList{
		Total: int(count),
		List:  list,
	}, nil
}

// PaySuccess 支付成功回调
func (s *Service) PaySuccess(ctx context.Context, w http.ResponseWriter, request *http.Request) error {
	ctx, span := tracer().Start(ctx, "ShopService.PaySuccess")
	defer span.End()

	// 验证并解析支付通知
	notifyData, err := s.VerifyAndParsePayNotify(ctx, w, request)
	if err != nil {
		span.RecordError(err)
		return err
	}

	// 处理订单支付
	if err := s.ProcessOrderPayment(ctx, notifyData); err != nil {
		span.RecordError(err)
		slog.Error("[Shop] ProcessOrderPayment failed", "err", err, "out_trade_no", notifyData.OutTradeNo)
		_ = PayResponse(w, "FAIL", "处理订单失败")
		return err
	}

	// 返回微信支付回调成功响应
	return PayResponse(w, "SUCCESS", "成功")
}

// VerifyAndParsePayNotify 验证并解析支付通知
func (s *Service) VerifyAndParsePayNotify(_ context.Context, w http.ResponseWriter, request *http.Request) (*PayNotifyData, error) {
	notify, err := wechat.V3ParseNotify(request)
	if err != nil {
		_ = PayResponse(w, "FAIL", "解析通知失败")
		return nil, err
	}

	notifyData := NewPayNotifyData(notify)

	if err := notifyData.VerifySign(s.WxMinPayService); err != nil {
		_ = PayResponse(w, "FAIL", "验签失败")
		return nil, err
	}

	if err := notifyData.Decrypt(); err != nil {
		_ = PayResponse(w, "FAIL", "解密失败")
		return nil, err
	}

	slog.Info("[Shop] PaySuccess", "notifyData", notifyData)

	// 验证交易状态
	if notifyData.TradeState != "SUCCESS" {
		_ = PayResponse(w, "FAIL", notifyData.TradeStateDesc)
		return nil, fmt.Errorf("交易状态异常: %s", notifyData.TradeStateDesc)
	}

	return notifyData, nil
}

// ProcessOrderPayment 处理订单支付
func (s *Service) ProcessOrderPayment(ctx context.Context, notifyData *PayNotifyData) error {
	return query.Q.Transaction(func(tx *query.Query) error {
		// 更新订单状态
		result, err := tx.Order.
			Where(tx.Order.OutTradeNo.Eq(notifyData.OutTradeNo),
				tx.Order.Status.Eq(string(model.OrderStatusPending))).
			Updates(map[string]any{
				tx.Order.Status.ColumnName().String():        model.OrderStatusPaid.String(),
				tx.Order.TransactionID.ColumnName().String(): notifyData.TransactionId,
			})
		if err != nil {
			return fmt.Errorf("更新订单状态失败: %w", err)
		}

		// 如果影响行数为0，说明订单已处理过，直接返回成功（幂等性）
		if result.RowsAffected == 0 {
			slog.Info("[Shop] order already processed", "out_trade_no", notifyData.OutTradeNo)
			return nil
		}

		// 查询订单详情，获取商品信息和设备ID
		order, err := tx.Order.
			Where(tx.Order.OutTradeNo.Eq(notifyData.OutTradeNo)).
			Preload(tx.Order.Goods).
			Preload(tx.Order.Goods.GoodsInfo).
			First()
		if err != nil {
			return fmt.Errorf("获取订单详情失败: %w", err)
		}

		// 开通会员：根据商品类型设置设备VIP过期时间
		if err := s.ActivateMembership(ctx, order, tx); err != nil {
			return fmt.Errorf("激活会员失败: %w", err)
		}

		return nil
	})
}

// ActivateMembership 激活会员
func (s *Service) ActivateMembership(ctx context.Context, order *model.Order, tx ...*query.Query) error {
	ctx, span := tracer().Start(ctx, "ShopService.activateMembership")
	defer span.End()

	if order.DeviceID == "" {
		return fmt.Errorf("订单缺少设备ID")
	}

	// 根据商品名称计算VIP过期时间
	var vipDuration time.Duration
	for _, goods := range order.Goods {
		switch goods.GoodsName {
		case "月卡":
			vipDuration = 30 * 24 * time.Hour // 30天
		case "年卡":
			vipDuration = 365 * 24 * time.Hour // 365天
		case "终身":
			vipDuration = 100 * 365 * 24 * time.Hour // 100年
		default:
			slog.Warn("[Shop] unknown goods name", "goods_name", goods.GoodsName)
			continue
		}
		break // 只处理第一个商品
	}

	if vipDuration == 0 {
		return fmt.Errorf("无法识别商品类型")
	}

	// 为指定设备设置VIP过期时间
	if err := s.DeviceRepo.SetDeviceVipExpireTime(ctx, order.DeviceID, vipDuration, tx...); err != nil {
		slog.Error("[Shop] SetDeviceVipExpireTime", "err", err, "device_id", order.DeviceID)
		return fmt.Errorf("设置设备VIP过期时间失败: %w", err)
	}

	slog.Info("[Shop] SetDeviceVipExpireTime success",
		"device_id", order.DeviceID,
		"duration", vipDuration.String(),
		"goods_name", order.Goods[0].GoodsName)

	return nil
}

// RefundNotify 退款回调
func (s *Service) RefundNotify(ctx context.Context, w http.ResponseWriter, request *http.Request) error {
	ctx, span := tracer().Start(ctx, "ShopService.RefundNotify")
	defer span.End()

	notify, err := wechat.V3ParseNotify(request)
	if err != nil {
		span.RecordError(err)
		_ = PayResponse(w, "FAIL", "解析通知失败")
		return err
	}

	notifyData := NewPayRefundNotifyData(notify)

	if err := notifyData.VerifySign(s.WxMinPayService); err != nil {
		span.RecordError(err)
		_ = PayResponse(w, "FAIL", "验签失败")
		return err
	}

	if err := notifyData.Decrypt(); err != nil {
		span.RecordError(err)
		_ = PayResponse(w, "FAIL", "解密失败")
		return err
	}

	slog.Info("[Shop] RefundNotify", "notifyData", notifyData)

	// 验证退款状态
	if notifyData.RefundStatus != "SUCCESS" {
		err := fmt.Errorf("退款状态异常: %s", notifyData.RefundStatus)
		span.RecordError(err)
		_ = PayResponse(w, "FAIL", notifyData.RefundStatus)
		return err
	}

	// 更新订单状态为已退款
	if err := s.MemberShopRepository.UpdateOrderStatusToRefunded(ctx, notifyData.OutTradeNo); err != nil {
		span.RecordError(err)
		_ = PayResponse(w, "FAIL", "更新订单状态失败")
		return err
	}

	// 返回微信支付回调成功响应
	return PayResponse(w, "SUCCESS", "成功")
}
