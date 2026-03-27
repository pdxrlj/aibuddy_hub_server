package wxmini

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

// WechatMiniProgram 微信小程序基础服务
type WechatMiniProgram struct {
	AppID     string
	AppSecret string
}

// GetJsCode2SessionResponse 微信登录凭证校验响应
type GetJsCode2SessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
	Rid        string `json:"rid"`
}

// NewWechatMiniProgram 创建微信小程序基础服务实例
func NewWechatMiniProgram(appID, appSecret string) *WechatMiniProgram {
	return &WechatMiniProgram{
		AppID:     appID,
		AppSecret: appSecret,
	}
}

// GetJsCode2Session 登录凭证校验，通过 code 获取用户 openid 和 session_key
func (w *WechatMiniProgram) GetJsCode2Session(code string) (*GetJsCode2SessionResponse, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code", w.AppID, w.AppSecret, code)
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "get wx js code 2 session")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read resp.Body")
	}

	var result GetJsCode2SessionResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal js code 2 session response")
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("errcode: %d, errmsg: %s", result.ErrCode, result.ErrMsg)
	}

	return &result, nil
}
