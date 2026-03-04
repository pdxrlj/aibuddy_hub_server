package baidu

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

// Signature 百度云API签名器
type Signature struct {
	AccessKeyID     string // AK
	SecretAccessKey string // SK
}

// AuthResult 认证结果
type AuthResult struct {
	Authorization string // 完整的认证字符串
	Timestamp     string // 时间戳 (x-bce-date 格式)
	SignedHeaders string // 签名的Header列表
}

// NewSignature 创建签名器
func NewSignature(accessKeyID, secretAccessKey string) *Signature {
	return &Signature{
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
	}
}

// GenerateAuth 生成认证字符串
func (s *Signature) GenerateAuth(method, path string, query url.Values, headers map[string]string) *AuthResult {
	ts := time.Now().UTC().Format("2006-01-02T15:04:05Z")

	// 确保headers中有x-bce-date
	if headers == nil {
		headers = make(map[string]string)
	}
	if _, ok := headers["x-bce-date"]; !ok {
		headers["x-bce-date"] = ts
	}

	// 生成CanonicalRequest
	canonicalRequest := s.buildCanonicalRequest(method, path, query, headers)

	// 生成SigningKey
	signingKeyStr := fmt.Sprintf("bce-auth-v1/%s/%s/1800", s.AccessKeyID, ts)
	signingKey := hmacSHA256Hex(signingKeyStr, s.SecretAccessKey)

	// 生成Signature
	signature := hmacSHA256Hex(canonicalRequest, signingKey)

	// 获取签名的Header列表
	signedHeaders := s.getSignedHeaders(headers)

	return &AuthResult{
		Authorization: fmt.Sprintf("%s/%s/%s", signingKeyStr, signedHeaders, signature),
		Timestamp:     ts,
		SignedHeaders: signedHeaders,
	}
}

// buildCanonicalRequest 构建CanonicalRequest
func (s *Signature) buildCanonicalRequest(method, path string, query url.Values, headers map[string]string) string {
	return fmt.Sprintf("%s\n%s\n%s\n%s",
		strings.ToUpper(method),
		s.buildCanonicalURI(path),
		s.buildCanonicalQueryString(query),
		s.buildCanonicalHeaders(headers))
}

// buildCanonicalURI 构建CanonicalURI
func (s *Signature) buildCanonicalURI(path string) string {
	if path == "" || path == "/" {
		return "/"
	}

	parts := strings.Split(path, "/")
	var encodedParts []string
	for _, part := range parts {
		if part == "" {
			continue
		}
		encodedParts = append(encodedParts, normalize(part, true))
	}

	return "/" + strings.Join(encodedParts, "/")
}

// buildCanonicalQueryString 构建CanonicalQueryString
func (s *Signature) buildCanonicalQueryString(query url.Values) string {
	if len(query) == 0 {
		return ""
	}

	var parts []string
	for key, values := range query {
		if strings.ToLower(key) == "authorization" {
			continue
		}
		for _, value := range values {
			parts = append(parts, fmt.Sprintf("%s=%s", normalize(key, true), normalize(value, true)))
		}
	}

	sort.Strings(parts)
	return strings.Join(parts, "&")
}

// buildCanonicalHeaders 构建CanonicalHeaders
func (s *Signature) buildCanonicalHeaders(headers map[string]string) string {
	defaultHeaders := []string{"host", "content-length", "content-type", "content-md5"}

	var signedHeaders []string
	headerMap := make(map[string]string)

	for key, value := range headers {
		lowerKey := strings.ToLower(key)

		isDefault := false
		for _, h := range defaultHeaders {
			if lowerKey == h {
				isDefault = true
				break
			}
		}

		if isDefault || strings.HasPrefix(lowerKey, "x-bce-") {
			if trimmedValue := strings.TrimSpace(value); trimmedValue != "" {
				signedHeaders = append(signedHeaders, lowerKey)
				headerMap[lowerKey] = trimmedValue
			}
		}
	}

	var lines = make([]string, 0, len(signedHeaders))
	for _, key := range signedHeaders {
		lines = append(lines, fmt.Sprintf("%s:%s", normalize(key, true), normalize(headerMap[key], true)))
	}

	sort.Strings(lines)
	return strings.Join(lines, "\n")
}

// getSignedHeaders 获取签名的Header列表
func (s *Signature) getSignedHeaders(headers map[string]string) string {
	defaultHeaders := []string{"host", "content-length", "content-type", "content-md5"}

	var signedHeaders []string
	for key := range headers {
		lowerKey := strings.ToLower(key)

		isDefault := false
		for _, h := range defaultHeaders {
			if lowerKey == h {
				isDefault = true
				break
			}
		}

		if isDefault || strings.HasPrefix(lowerKey, "x-bce-") {
			if trimmedValue := strings.TrimSpace(headers[key]); trimmedValue != "" {
				signedHeaders = append(signedHeaders, lowerKey)
			}
		}
	}

	sort.Strings(signedHeaders)

	var encodedKeys = make([]string, 0, len(signedHeaders))
	for _, key := range signedHeaders {
		encodedKeys = append(encodedKeys, normalize(key, true))
	}

	return strings.Join(encodedKeys, ";")
}

// normalize URL编码（与Postman版本一致）
func normalize(input string, _ bool) string {
	if input == "" {
		return ""
	}

	result := url.QueryEscape(input)
	result = strings.ReplaceAll(result, "+", "%20")
	result = strings.ReplaceAll(result, "%21", "!")
	result = strings.ReplaceAll(result, "%27", "'")
	result = strings.ReplaceAll(result, "%28", "(")
	result = strings.ReplaceAll(result, "%29", ")")
	result = strings.ReplaceAll(result, "%2A", "*")
	// encodingSlash always true, so we never unescape %2F

	return result
}

// hmacSHA256Hex 计算HMAC-SHA256并返回十六进制字符串
func hmacSHA256Hex(data, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
