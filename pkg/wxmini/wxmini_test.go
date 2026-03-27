package wxmini

import (
	"aibuddy/pkg/config"
	"aibuddy/pkg/helpers"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var _wxMini *WechatMiniProgram

func TestMain(m *testing.M) {
	config.Setup()
	wchatConfig := config.Instance.Wechat
	helpers.PP(wchatConfig)
	_wxMini = NewWechatMiniProgram(wchatConfig.AppID, wchatConfig.AppSecret)
	os.Exit(m.Run())
}

func TestGetJsCode2Session(t *testing.T) {
	const code = "0f1RAeGa1ipqtK07bVGa1y1CVA1RAeG9"

	resp, err := _wxMini.GetJsCode2Session(code)
	assert.NoError(t, err)
	// {
	//   	"openid": "oNIUD7MTW4EXrR_Toae6c-8EdLak",
	//   	"session_key": "4i4Tam9JQl4wb6GPet9LHg==",
	//   	"unionid": "olmjS68UZvMdpGWoZ45KjJ_vjryc",
	//   	"errcode": 0,
	//   	"errmsg": "",
	//   	"rid": ""
	// }
	assert.Equal(t, 0, resp.ErrCode)
	assert.NotEmpty(t, resp.OpenID) // oNIUD7MTW4EXrR_Toae6c-8EdLak
	assert.NotEmpty(t, resp.SessionKey)
	assert.NotEmpty(t, resp.UnionID)

	helpers.PP(resp)
}
