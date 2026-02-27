// Package auth 用户认证
package auth

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
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
	SecretKey:     "your-secret-key-here", // 应该从配置文件读取
	TokenExpiry:   24 * time.Hour,         // 24小时过期
	RefreshExpiry: 7 * 24 * time.Hour,     // 刷新令牌7天过期
	Issuer:        "wechat-backend",
}

// GenerateToken 生成 JWT 令牌
func GenerateToken(uid int64, phone string, openID string) (string, error) {
	claims := &MyClaims{
		UID:       uid,
		Phone:     phone,
		OpenID:    openID,
		ExpiresAt: time.Now().Add(jwtConfig.TokenExpiry).Unix(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtConfig.TokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    jwtConfig.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtConfig.SecretKey))
}

// ValidateToken 验证JWT令牌
func ValidateToken(tokenString string) (*MyClaims, error) {
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
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// RefreshToken 刷新令牌
func RefreshToken(oldTokenString string) (string, error) {
	// 验证旧令牌（即使过期也要验证签名）
	token, err := jwt.ParseWithClaims(oldTokenString, &MyClaims{}, func(_ *jwt.Token) (interface{}, error) {
		return []byte(jwtConfig.SecretKey), nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to parse old token: %w", err)
	}

	claims, ok := token.Claims.(*MyClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	// 检查令牌是否在刷新有效期内
	if time.Now().Unix() > claims.IssuedAt+int64(jwtConfig.RefreshExpiry.Seconds()) {
		return "", fmt.Errorf("refresh token has expired")
	}

	// 生成新令牌
	return GenerateToken(claims.UID, claims.Phone, claims.OpenID)
}

// ValidateSimpleToken 简单的令牌验证函数
func ValidateSimpleToken(token, expectedToken string) bool {
	return strings.EqualFold(token, expectedToken)
}
