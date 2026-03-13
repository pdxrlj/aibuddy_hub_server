// Package ahttp 提供 HTTP 框架封装
package ahttp

import (
	"aibuddy/pkg/buddyerror"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// State 请求状态
type State struct {
	State any
	Ctx   echo.Context
}

// Base HTTP 基础结构
type Base struct {
	*echo.Echo
	State *State
}

// NewBase 创建 HTTP 基础结构
func NewBase(echo *echo.Echo, state *State) *Base {
	return &Base{
		Echo:  echo,
		State: state,
	}
}

// GET 注册 GET 路由
func (b *Base) GET(path string, handler any, middlewares ...echo.MiddlewareFunc) {
	b.Add(http.MethodGet, path, handler, middlewares...)
}

// POST 注册 POST 路由
func (b *Base) POST(path string, handler any, middlewares ...echo.MiddlewareFunc) {
	b.Add(http.MethodPost, path, handler, middlewares...)
}

// PUT 注册 PUT 路由
func (b *Base) PUT(path string, handler any, middlewares ...echo.MiddlewareFunc) {
	b.Add(http.MethodPut, path, handler, middlewares...)
}

// PATCH 注册 PATCH 路由
func (b *Base) PATCH(path string, handler any, middlewares ...echo.MiddlewareFunc) {
	b.Add(http.MethodPatch, path, handler, middlewares...)
}

// OPTIONS 注册 OPTIONS 路由
func (b *Base) OPTIONS(path string, handler any, middlewares ...echo.MiddlewareFunc) {
	b.Add(http.MethodOptions, path, handler, middlewares...)
}

// HEAD 注册 HEAD 路由
func (b *Base) HEAD(path string, handler any, middlewares ...echo.MiddlewareFunc) {
	b.Add(http.MethodHead, path, handler, middlewares...)
}

// DELETE 注册 DELETE 路由
func (b *Base) DELETE(path string, handler any, middlewares ...echo.MiddlewareFunc) {
	b.Add(http.MethodDelete, path, handler, middlewares...)
}

// Group HTTP 路由组
type Group struct {
	EchoGroup *echo.Group
	*Base
}

// Group 创建路由组
func (b *Base) Group(path string, middlewares []echo.MiddlewareFunc, handler func(group *Group)) *echo.Group {
	group := b.Echo.Group(path, middlewares...)
	g := &Group{
		EchoGroup: group,
		Base:      b,
	}
	handler(g)
	return group
}

// Group 在 Group 中创建子路由组
func (g *Group) Group(path string, middlewares []echo.MiddlewareFunc, handler func(group *Group)) *echo.Group {
	subGroup := g.EchoGroup.Group(path, middlewares...)
	gg := &Group{
		EchoGroup: subGroup,
		Base:      g.Base,
	}
	handler(gg)
	return subGroup
}

// GET 注册 GET 路由
func (g *Group) GET(path string, handler any, middlewares ...echo.MiddlewareFunc) {
	g.Add(http.MethodGet, path, handler, middlewares...)
}

// POST 注册 POST 路由
func (g *Group) POST(path string, handler any, middlewares ...echo.MiddlewareFunc) {
	g.Add(http.MethodPost, path, handler, middlewares...)
}

// PUT 注册 PUT 路由
func (g *Group) PUT(path string, handler any, middlewares ...echo.MiddlewareFunc) {
	g.Add(http.MethodPut, path, handler, middlewares...)
}

// PATCH 注册 PATCH 路由
func (g *Group) PATCH(path string, handler any, middlewares ...echo.MiddlewareFunc) {
	g.Add(http.MethodPatch, path, handler, middlewares...)
}

// OPTIONS 注册 OPTIONS 路由
func (g *Group) OPTIONS(path string, handler any, middlewares ...echo.MiddlewareFunc) {
	g.Add(http.MethodOptions, path, handler, middlewares...)
}

// HEAD 注册 HEAD 路由
func (g *Group) HEAD(path string, handler any, middlewares ...echo.MiddlewareFunc) {
	g.Add(http.MethodHead, path, handler, middlewares...)
}

// DELETE 注册 DELETE 路由
func (g *Group) DELETE(path string, handler any, middlewares ...echo.MiddlewareFunc) {
	g.Add(http.MethodDelete, path, handler, middlewares...)
}

// Add 添加路由到 Group
func (g *Group) Add(method, path string, handler any, middlewares ...echo.MiddlewareFunc) {
	handlerValue, inputNumber := validateHandler(handler)
	g.EchoGroup.Add(method, path, g.createHandler(handlerValue, inputNumber), middlewares...)
}

// Add 添加路由
// handler 函数签名: func(state *State, request any) error
func (b *Base) Add(method, path string, handler any, middlewares ...echo.MiddlewareFunc) {
	handlerValue, inputNumber := validateHandler(handler)
	b.Echo.Add(method, path, b.createHandler(handlerValue, inputNumber), middlewares...)
}

// validateHandler 验证 handler 函数签名
func validateHandler(handler any) (reflect.Value, int) {
	handlerValue := reflect.ValueOf(handler)
	if handlerValue.Kind() != reflect.Func {
		panic("参数handler必须是一个函数")
	}
	handlerType := handlerValue.Type()
	if handlerType.NumIn() < 1 {
		panic("参数handler必须至少有一个参数")
	}
	if handlerType.In(0) != reflect.TypeOf(&State{}) {
		panic("参数handler第一个参数必须是State")
	}

	if handlerType.NumOut() != 1 || handlerType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
		panic("参数handler必须返回一个error")
	}

	return handlerValue, handlerType.NumIn()
}

// SkipBodyBinder 跳过 body 绑定的接口
// 实现 this 接口的请求结构体将不会尝试从 body 绑定数据
type SkipBodyBinder interface {
	SkipBodyBind()
}

// createHandler 创建 echo.HandlerFunc
func (b *Base) createHandler(handlerValue reflect.Value, inputNumber int) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		if inputNumber == 1 {
			return b.ResoverHandler(ctx, handlerValue)
		}
		// 获取 handler 原始参数类型
		handlerParamType := handlerValue.Type().In(1)
		// 如果参数是指针类型，解引用获取基础类型用于创建实例
		requestType := handlerParamType
		if requestType.Kind() == reflect.Ptr {
			requestType = requestType.Elem()
		}

		instance := reflect.New(requestType).Interface()

		if _, ok := instance.(SkipBodyBinder); ok {
			if err := bindPathAndQuery(ctx, instance); err != nil {
				return NewResponse(ctx).
					SetStatus(buddyerror.GetBuddyErrorCode(buddyerror.ErrParamError)).
					SetMessage("参数格式错误").
					Error(err)
			}
		} else if err := ctx.Bind(instance); err != nil {
			msg := "参数格式错误"
			var typeErr *json.UnmarshalTypeError
			if errors.As(err, &typeErr) {
				msg = fmt.Sprintf("参数 %s 类型错误，期望 %s 类型", typeErr.Field, typeErr.Type.Name())
			}
			return NewResponse(ctx).
				SetStatus(buddyerror.GetBuddyErrorCode(buddyerror.ErrParamError)).
				SetMessage(msg).
				Error(errors.New(msg))
		}

		if err := b.RequestValidator(ctx, instance, requestType); err != nil {
			return err
		}

		// 验证失败时响应已被写入
		if ctx.Response().Committed {
			return nil
		}

		// 如果 handler 期望指针，则传指针；否则传值
		if handlerParamType.Kind() == reflect.Ptr {
			return b.ResoverHandler(ctx, handlerValue, instance)
		}
		// 解引用获取值
		return b.ResoverHandler(ctx, handlerValue, reflect.ValueOf(instance).Elem().Interface())
	}
}

// RequestValidator 验证请求参数
func (b *Base) RequestValidator(ctx echo.Context, requestType any, request reflect.Type) error {
	if b.Validator != nil {
		if err := ctx.Validate(requestType); err != nil {
			if validationErrors, ok := err.(*ValidationErrors); ok {
				// 优先使用 msg tag 中的自定义消息
				customMsg := validationErrors.GetFirstErrorMsg()
				if customMsg != "" {
					return NewResponse(ctx).SetStatus(buddyerror.GetBuddyErrorCode(buddyerror.ErrParamError)).
						SetMessage(customMsg).
						Error(errors.New(customMsg))
				}

				// 没有 msg tag，使用默认逻辑
				firstError := validationErrors.Errors[0]
				fieldType := firstError.StructField()
				actualTag := firstError.ActualTag()

				validationFieldTag, validationRule := getValidationFieldTag(request, fieldType, actualTag)

				err = fmt.Errorf("参数 %s 无法通过 %s 规则的验证", validationFieldTag, validationRule)
				return NewResponse(ctx).SetStatus(buddyerror.GetBuddyErrorCode(buddyerror.ErrParamError)).
					SetMessage(buddyerror.ErrParamError.Error()).
					Error(err)
			}

			return NewResponse(ctx).
				SetStatus(buddyerror.GetBuddyErrorCode(buddyerror.ErrParamError)).
				SetMessage(buddyerror.ErrParamError.Error()).Error(err)
		}
	}
	return nil
}

// ResoverHandler 处理请求
func (b *Base) ResoverHandler(ctx echo.Context, handlerValue reflect.Value, requests ...any) error {
	state := &State{
		Ctx:   ctx,
		State: b.State,
	}

	in := []reflect.Value{
		reflect.ValueOf(state),
	}
	if len(requests) > 0 {
		in = append(in, reflect.ValueOf(requests[0]))
	}
	result := handlerValue.Call(in)[0]
	if result.IsNil() {
		return nil
	}
	if respError, ok := result.Interface().(error); ok && respError != nil {
		slog.Error("处理函数返回数据异常", "error", respError)
	}

	return nil
}

// Validator 参数验证器
type Validator struct {
	Validator *validator.Validate
}

// ValidationErrors 验证错误
type ValidationErrors struct {
	Errors     []validator.FieldError `json:"errors"`
	StructType reflect.Type           // 保存结构体类型用于获取 msg tag
}

// Error 实现 error 接口
func (v *ValidationErrors) Error() string {
	errs := make([]string, 0, len(v.Errors))
	for _, err := range v.Errors {
		// 使用翻译器获取中文错误消息
		if Trans != nil {
			errs = append(errs, err.Translate(Trans))
		} else {
			errs = append(errs, err.Error())
		}
	}
	return strings.Join(errs, ", ")
}

// GetFirstErrorMsg 获取第一个错误的 msg tag 消息，优先使用自定义消息
func (v *ValidationErrors) GetFirstErrorMsg() string {
	if len(v.Errors) == 0 {
		return ""
	}

	firstError := v.Errors[0]
	structField := firstError.StructField()
	actualTag := firstError.ActualTag()

	// 尝试获取 msg tag
	if v.StructType != nil {
		structType := v.StructType
		if structType.Kind() == reflect.Ptr {
			structType = structType.Elem()
		}
		if structType.Kind() == reflect.Struct {
			field, ok := structType.FieldByName(structField)
			if ok {
				msgTag := field.Tag.Get("msg")
				if msgTag != "" {
					// 解析 msg tag: "required:不能为空|chmobile:手机号格式错误"
					if customMsg := parseMsgTag(msgTag, actualTag); customMsg != "" {
						return customMsg
					}
				}
			}
		}
	}

	// 没有自定义消息，使用翻译
	if Trans != nil {
		return firstError.Translate(Trans)
	}
	return firstError.Error()
}

// parseMsgTag 解析 msg tag，格式: "required:不能为空|chmobile:手机号格式错误"
func parseMsgTag(msgTag, actualTag string) string {
	rules := strings.Split(msgTag, "|")
	for _, rule := range rules {
		parts := strings.SplitN(rule, ":", 2)
		if len(parts) == 2 && parts[0] == actualTag {
			return parts[1]
		}
	}
	// 如果没有指定规则，整个 msg 作为消息
	if len(rules) == 1 && !strings.Contains(msgTag, ":") {
		return msgTag
	}
	return ""
}

// Validate 验证结构体
func (v *Validator) Validate(i interface{}) error {
	err := v.Validator.Struct(i)
	if err == nil {
		return nil
	}
	if validatorErrors, ok := err.(validator.ValidationErrors); ok {
		return &ValidationErrors{
			Errors:     validatorErrors,
			StructType: reflect.TypeOf(i),
		}
	}

	return err
}

func getValidationFieldTag(structType reflect.Type, defaultTag, actualTag string) (fieldTag, validationRule string) {
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}

	structField, _ := structType.FieldByName(defaultTag)

	var ok bool
	fieldTag, ok = structField.Tag.Lookup("query")
	if !ok {
		fieldTag, ok = structField.Tag.Lookup("json")
	}
	if !ok {
		fieldTag, ok = structField.Tag.Lookup("path")
	}
	if !ok {
		fieldTag, ok = structField.Tag.Lookup("form")
	}
	if !ok {
		fieldTag, ok = structField.Tag.Lookup("header")
	}
	if !ok {
		fieldTag, ok = structField.Tag.Lookup("param")
	}

	if !ok {
		fieldTag = defaultTag
	}

	// 获取验证规则
	validationRule = structField.Tag.Get("validate")
	// 获取 validationRule 中的 actualTag
	rules := strings.Split(validationRule, ",")
	for _, rule := range rules {
		if strings.HasPrefix(rule, actualTag) {
			validationRule = rule
			break
		}
	}

	return fieldTag, validationRule
}

// Resposne 返回响应对象（注意：方法名拼写错误，保持兼容性）
func (s *State) Resposne() *Response {
	return NewResponse(s.Ctx)
}

// Context 获取请求上下文
func (s *State) Context() context.Context {
	return s.Ctx.Request().Context()
}

// bindPathAndQuery 只绑定 path 和 query 参数
func bindPathAndQuery(ctx echo.Context, i interface{}) error {
	b := new(echo.DefaultBinder)
	if err := b.BindQueryParams(ctx, i); err != nil {
		return err
	}
	if err := b.BindPathParams(ctx, i); err != nil {
		return err
	}
	return nil
}
