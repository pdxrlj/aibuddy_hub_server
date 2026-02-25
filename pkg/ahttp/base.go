// Package ahttp 提供 HTTP 框架封装
package ahttp

import (
	"aibuddy/pkg/buddyerror"
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

// Add 添加路由
// handler 函数签名: func(state *State, request any) error
func (b *Base) Add(method, path string, handler any, middlewares ...echo.MiddlewareFunc) {
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

	inputNumber := handlerType.NumIn()

	b.Echo.Add(method, path, func(ctx echo.Context) error {
		if inputNumber == 1 {
			return b.ResoverHandler(ctx, handlerValue)
		}
		request := handlerType.In(1)
		if request.Kind() == reflect.Ptr {
			request = request.Elem()
		}

		requestType := reflect.New(request).Interface()

		if err := ctx.Bind(requestType); err != nil {
			return NewResponse(ctx).
				SetStatus(buddyerror.GetBuddyErrorCode(buddyerror.ErrParamError)).
				SetMessage(buddyerror.ErrParamError.Error()).
				Error(err)
		}

		if err := b.RequestValidator(ctx, requestType, request); err != nil {
			return err
		}

		return b.ResoverHandler(ctx, handlerValue, requestType)
	}, middlewares...)
}

// RequestValidator 验证请求参数
func (b *Base) RequestValidator(ctx echo.Context, requestType any, request reflect.Type) error {
	if b.Validator != nil {
		if err := ctx.Validate(requestType); err != nil {
			if validationErrors, ok := err.(*ValidationErrors); ok {
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
	Errors []validator.FieldError `json:"errors"`
}

// Error 实现 error 接口
func (v *ValidationErrors) Error() string {
	errs := make([]string, 0, len(v.Errors))
	for _, err := range v.Errors {
		errs = append(errs, err.Error())
	}
	return strings.Join(errs, ", ")
}

// Validate 验证结构体
func (v *Validator) Validate(i interface{}) error {
	err := v.Validator.Struct(i)
	if err == nil {
		return nil
	}
	if validatorErrors, ok := err.(validator.ValidationErrors); ok {
		return &ValidationErrors{
			Errors: validatorErrors,
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
