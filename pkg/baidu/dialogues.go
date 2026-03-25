// Package baidu 百度云API dialogues 接口
package baidu

import (
	"fmt"
	"log/slog"
	"net/url"

	"aibuddy/pkg/config"
)

// Dialogues 对话记录API
type Dialogues struct {
	client *Client
}

// NewDialogues 创建对话记录客户端
func NewDialogues() *Dialogues {
	return &Dialogues{client: NewClient()}
}

// DialogueItem 对话记录项
type DialogueItem struct {
	Type      string `json:"type"`      // 文本类型：QUESTION 或 ANSWER
	Timestamp int64  `json:"timestamp"` // 时间戳
	Text      string `json:"text"`      // 对话内容
}

// DialoguesResponse 对话记录响应
type DialoguesResponse struct {
	PageNo   int            `json:"pageNo"`   // 当前页
	PageSize int            `json:"pageSize"` // 当前页查询到的数量
	Data     []DialogueItem `json:"data"`     // 对话记录列表
}

// DialoguesRequest 对话记录请求参数
type DialoguesRequest struct {
	AppID     string `json:"appId"`     // 互动应用ID
	UserID    string `json:"userId"`    // 业务侧用户唯一ID
	PageNo    int    `json:"pageNo"`    // 起始页
	PageSize  int    `json:"pageSize"`  // 每页最大值（不超过100）
	BeginTime int64  `json:"beginTime"` // 开始时间，单位秒，包含
	EndTime   int64  `json:"endTime"`   // 结束时间，单位秒，不包含
}

// GetDialogues 获取对话记录
func (d *Dialogues) GetDialogues(req *DialoguesRequest) (*DialoguesResponse, error) {
	if req.AppID == "" {
		req.AppID = config.Instance.Baidu.AppID
	}
	slog.Info("[Baidu] GetDialogues", "开始时间", req.BeginTime, "结束时间", req.EndTime)
	path := "/api/v1/dialogues"
	query := url.Values{}
	query.Set("appId", req.AppID)
	query.Set("userId", req.UserID)
	query.Set("pageNo", fmt.Sprintf("%d", req.PageNo))
	query.Set("pageSize", fmt.Sprintf("%d", req.PageSize))
	query.Set("beginTime", fmt.Sprintf("%d", req.BeginTime))
	query.Set("endTime", fmt.Sprintf("%d", req.EndTime))

	var result DialoguesResponse
	if err := d.client.Request("GET", path, query, nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
