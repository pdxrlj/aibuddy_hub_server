// Package auth 用户认证
package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
)

// MyClaims 用户认证结构体
type MyClaims struct {
	UID       int64  `json:"uid"`
	Phone     string `json:"phone"`
	OpenID    string `json:"open_id,omitempty"`
	ExpiresAt int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
	jwt.RegisteredClaims
}

// JWTConfig 定义 JWT 配置结构
type JWTConfig struct {
	SecretKey     string
	TokenExpiry   time.Duration
	RefreshExpiry time.Duration
	Issuer        string
}

var jwtConfig = &JWTConfig{
	SecretKey:     "aibuddy",          // 应该从配置文件读取
	TokenExpiry:   24 * time.Hour,     // 24小时过期
	RefreshExpiry: 7 * 24 * time.Hour, // 刷新令牌7天过期
	Issuer:        "wechat-backend",
}

// GenerateToken 生成 JWT 令牌
func GenerateToken(uid int64, phone string, openID string) (string, int64, error) {
	expiresTime := time.Now().Add(jwtConfig.TokenExpiry)
	claims := &MyClaims{
		UID:       uid,
		Phone:     phone,
		OpenID:    openID,
		ExpiresAt: expiresTime.Unix(),
		IssuedAt:  time.Now().Unix(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    jwtConfig.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	result, err := token.SignedString([]byte(jwtConfig.SecretKey))
	return result, expiresTime.Unix(), err
}

// ValidateToken 验证JWT令牌
func ValidateToken(c echo.Context, tokenString string) (*MyClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtConfig.SecretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		// 检查令牌是否过期
		if time.Now().Unix() > claims.ExpiresAt {
			return nil, fmt.Errorf("token has expired")
		}
		c.Set("uid", claims.UID)
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// RefreshToken 刷新令牌
func RefreshToken(ctx context.Context, oldTokenString string) (string, int64, error) {
	_, span := tracer().Start(ctx, "RefreshToken")
	defer span.End()

	// 验证旧令牌（即使过期也要验证签名）
	token, err := jwt.ParseWithClaims(oldTokenString, &MyClaims{}, func(_ *jwt.Token) (interface{}, error) {
		return []byte(jwtConfig.SecretKey), nil
	})

	if err != nil {
		span.RecordError(err)
		return "", 0, fmt.Errorf("failed to parse old token: %w", err)
	}

	claims, ok := token.Claims.(*MyClaims)
	if !ok {
		span.RecordError(errors.New("invalid token claims"))
		return "", 0, fmt.Errorf("invalid token claims")
	}

	// 检查令牌是否在刷新有效期内
	if time.Now().Unix() > claims.IssuedAt+int64(jwtConfig.RefreshExpiry.Seconds()) {
		span.RecordError(errors.New("refresh token has expired"))
		return "", 0, fmt.Errorf("refresh token has expired")
	}

	// 生成新令牌
	return GenerateToken(claims.UID, claims.Phone, claims.OpenID)
}

// ValidateSimpleToken 简单的令牌验证函数
func ValidateSimpleToken(token, expectedToken string) bool {
	return strings.EqualFold(token, expectedToken)
}

// GetUIDFromContext 通过Context获取用户uid
func GetUIDFromContext(c echo.Context) (int64, error) {
	userID := c.Get("uid")

	uid, ok := userID.(int64)
	if !ok {
		return 0, errors.New("invalid user ID type")
	}

	return uid, nil
}
