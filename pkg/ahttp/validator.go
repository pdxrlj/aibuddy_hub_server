// Package ahttp 提供 HTTP 框架封装
package ahttp

import (
	logger "aibuddy/pkg/log"
	"log/slog"
	"reflect"
	"regexp"

	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

// Trans 全局翻译器
var Trans ut.Translator

// 预编译正则表达式
var (
	mobileRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)
	macRegex    = regexp.MustCompile(`^([0-9A-Fa-f]{2}:){5}([0-9A-Fa-f]{2})$`)
)

// NewValidator 创建验证器
func NewValidator() *validator.Validate {
	v := validator.New()

	// 初始化中文翻译器
	zhLocale := zh.New()
	uni := ut.New(zhLocale, zhLocale)
	Trans, _ = uni.GetTranslator("zh")

	// 注册默认中文翻译
	if err := zh_translations.RegisterDefaultTranslations(v, Trans); err != nil {
		slog.Error(logger.ValidateRegister, "error", err)
	}

	// 注册自定义验证规则
	if err := v.RegisterValidation("chmobile", validateMobile); err != nil {
		slog.Error(logger.ValidateRegister, "error", err)
	}
	if err := v.RegisterValidation("aimac", validateMAC); err != nil {
		slog.Error(logger.ValidateRegister, "error", err)
	}
	if err := v.RegisterValidation("required_if_gt", validateRequiredIfGT); err != nil {
		slog.Error(logger.ValidateRegister, "error", err)
	}

	if err := v.RegisterValidation("ota_required", validateOTA); err != nil {
		slog.Error(logger.ValidateRegister, "error", err)
	}

	// 注册自定义翻译
	registerCustomTranslations(v)

	return v
}

// registerCustomTranslations 注册自定义错误消息
func registerCustomTranslations(v *validator.Validate) {
	// 自定义 required 消息
	if err := v.RegisterTranslation("required", Trans, func(ut ut.Translator) error {
		return ut.Add("required", "{0}不能为空", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("required", fe.Field())
		return t
	}); err != nil {
		slog.Error(logger.ValidateRegister, "error", err)
	}

	// 自定义 chmobile 消息
	if err := v.RegisterTranslation("chmobile", Trans, func(ut ut.Translator) error {
		return ut.Add("chmobile", "{0}必须是有效的手机号码", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("chmobile", fe.Field())
		return t
	}); err != nil {
		slog.Error(logger.ValidateRegister, "error", err)
	}

	// 自定义 aimac 消息
	if err := v.RegisterTranslation("aimac", Trans, func(ut ut.Translator) error {
		return ut.Add("aimac", "{0}必须是有效的MAC地址", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("aimac", fe.Field())
		return t
	}); err != nil {
		slog.Error(logger.ValidateRegister, "error", err)
	}

	// 自定义 required_if_gt 消息
	if err := v.RegisterTranslation("required_if_gt", Trans, func(ut ut.Translator) error {
		return ut.Add("required_if_gt", "{0}必须大于0", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("required_if_gt", fe.Field())
		return t
	}); err != nil {
		slog.Error(logger.ValidateRegister, "error", err)
	}

	// 自定义 ota_required 消息
	if err := v.RegisterTranslation("ota_required", Trans, func(ut ut.Translator) error {
		return ut.Add("ota_required", "send_all与device_ids不能同时存在", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("ota_required", fe.Field())
		return t
	}); err != nil {
		slog.Error(logger.ValidateRegister, "error", err)
	}
}

// validateMobile 手机号验证
func validateMobile(fl validator.FieldLevel) bool {
	mobile := fl.Field().String()
	return mobileRegex.MatchString(mobile)
}

// validateMAC MAC地址验证（支持字符串和字符串切片）
func validateMAC(fl validator.FieldLevel) bool {
	field := fl.Field()

	// 处理字符串切片
	if field.Kind().String() == "slice" {
		for i := 0; i < field.Len(); i++ {
			mac := field.Index(i).String()
			if !macRegex.MatchString(mac) {
				return false
			}
		}
		return true
	}

	// 处理单个字符串
	mac := field.String()
	return macRegex.MatchString(mac)
}

// validateRequiredIfGT 当条件字段等于指定值时，当前字段必须大于0
// 参数格式: field=value 例如: Fmt=voice 表示当 Fmt=voice 时，当前字段必须>0
func validateRequiredIfGT(fl validator.FieldLevel) bool {
	params := fl.Param()
	if params == "" {
		return true
	}

	// 解析参数 field=value
	parts := splitTwo(params, "=")
	if len(parts) != 2 {
		return false
	}

	conditionField := parts[0]
	conditionValue := parts[1]

	// 获取条件字段的值
	parent := fl.Parent()
	conditionFieldValue := parent.FieldByName(conditionField)
	if !conditionFieldValue.IsValid() {
		return true // 条件字段不存在，跳过验证
	}

	// 检查条件字段是否等于指定值
	if conditionFieldValue.String() != conditionValue {
		return true // 条件不满足，跳过验证
	}

	// 条件满足，检查当前字段是否大于0
	field := fl.Field()
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() > 0
	case reflect.Float32, reflect.Float64:
		return field.Float() > 0
	default:
		return true
	}
}

func splitTwo(s, sep string) []string {
	for i := 0; i < len(s); i++ {
		if s[i:i+1] == sep {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

// validateOTA 结构体验证：send_all 与 device_ids 不能同时存在
func validateOTA(fl validator.FieldLevel) bool {
	parent := fl.Parent()
	sendAllField := parent.FieldByName("SendAll")
	deviceIDsField := parent.FieldByName("DeviceIDs")

	if !sendAllField.IsValid() || !deviceIDsField.IsValid() {
		return true
	}

	sendAll := sendAllField.Bool()
	deviceIDs, _ := deviceIDsField.Interface().([]string)

	if sendAll && len(deviceIDs) > 0 {
		return false
	}

	return true
}
