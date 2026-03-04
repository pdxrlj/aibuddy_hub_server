// Package baidu 百度云API客户端
package baidu

import (
	"aibuddy/pkg/config"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-resty/resty/v2"
)

const baseURL = "https://rtc-aiagent.baidubce.com"

// Client 百度云API客户端
type Client struct {
	httpClient *resty.Client
	signer     *Signature
	host       string
}

// NewClient 创建百度云API客户端
func NewClient() *Client {
	cfg := config.Instance.Baidu
	return &Client{
		httpClient: resty.New().
			SetRetryCount(3).
			SetRetryWaitTime(500 * time.Millisecond).
			SetRetryMaxWaitTime(5 * time.Second).
			AddRetryCondition(func(r *resty.Response, err error) bool {
				if err != nil {
					return true
				}
				return r.StatusCode() >= http.StatusInternalServerError
			}),
		signer: NewSignature(cfg.Ak, cfg.Sk),
		host:   "rtc-aiagent.baidubce.com",
	}
}

// Request 发送带自动签名的请求
func (c *Client) Request(method, path string, query url.Values, body any, result any) error {
	auth := c.signer.GenerateAuth(method, path, query, map[string]string{
		"host": c.host,
	})

	req := c.buildRequest(auth, query, body)
	resp, err := c.executeRequest(req, method, baseURL+path)
	if err != nil {
		return err
	}

	return c.handleResponse(resp, result)
}

// buildRequest 构建请求
func (c *Client) buildRequest(auth *AuthResult, query url.Values, body any) *resty.Request {
	req := c.httpClient.R().
		SetHeader("Authorization", auth.Authorization).
		SetHeader("Content-Type", "application/json").
		SetHeader("x-bce-date", auth.Timestamp)

	if len(query) > 0 {
		req.SetQueryParamsFromValues(query)
	}

	if body != nil {
		req.SetBody(body)
	}

	return req
}

// executeRequest 执行请求
func (c *Client) executeRequest(req *resty.Request, method, urlStr string) (*resty.Response, error) {
	switch method {
	case "GET":
		return req.Get(urlStr)
	case "POST":
		return req.Post(urlStr)
	case "PUT":
		return req.Put(urlStr)
	case "DELETE":
		return req.Delete(urlStr)
	default:
		return nil, fmt.Errorf("unsupported method: %s", method)
	}
}

// handleResponse 处理响应
func (c *Client) handleResponse(resp *resty.Response, result any) error {
	if resp == nil {
		return fmt.Errorf("请求失败: 无响应")
	}

	if err := resp.Error(); err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}

	if resp.IsError() {
		return fmt.Errorf("请求错误: %s", resp.String())
	}

	if result != nil {
		if err := json.Unmarshal(resp.Body(), result); err != nil {
			return fmt.Errorf("解析响应失败: %w", err)
		}
	}

	return nil
}
