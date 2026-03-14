package shop

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// TokenURL 获取token url
	TokenURL = "https://api.weixin.qq.com/cgi-bin/token"
)

// AccessTokenResp token响应数据
type AccessTokenResp struct {
	ErrorCode   int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// GetAccessToken 获取商城的token
func GetAccessToken(_ context.Context, appID string, secret string) (AccessTokenResp, error) {
	var result AccessTokenResp
	uri := fmt.Sprintf("%s?grant_type=client_credential&appid=%s&secret=%s", TokenURL, appID, secret)

	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	response, err := client.Get(uri)
	if err != nil {
		return result, err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return result, err
	}
	if err := response.Body.Close(); err != nil {
		return result, err
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return result, err
	}

	return result, nil
}
