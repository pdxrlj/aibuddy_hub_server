// Package ahttp 提供 HTTP 框架封装
package ahttp

import (
	logger "aibuddy/pkg/log"
	"log/slog"
	"regexp"

	"github.com/go-playground/validator/v10"
)

// NewValidator 创建验证器
func NewValidator() *validator.Validate {
	validator := validator.New()

	if err := validator.RegisterValidation("chmobile", validateMobile); err != nil {
		slog.Error(logger.ValidateRegister, "error", err)
	}
	if err := validator.RegisterValidation("aimac", validateMAC); err != nil {
		slog.Error(logger.ValidateRegister, "error", err)
	}

	return validator
}

// validateMobile 手机号验证
func validateMobile(fl validator.FieldLevel) bool {
	mobile := fl.Field().String()
	matched, _ := regexp.MatchString(`^1[3-9]\d{9}$`, mobile)
	return matched
}

// validateMAC MAC地址验证
func validateMAC(fl validator.FieldLevel) bool {
	mac := fl.Field().String()
	matched, _ := regexp.MatchString(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`, mac)
	return matched
}
