// Package feedback defines the DTOs for the feedback.
package feedback

import "encoding/json"

// CreateFeedbackRequest 创建反馈请求
type CreateFeedbackRequest struct {
	FeedbackType string   `json:"feedback_type" form:"feedback_type" validate:"required,oneof=function_exception experience_problem product_suggestion other"`
	Content      string   `json:"content" form:"content" validate:"required,min=1,max=1000"`
	Images       []string `json:"images" form:"images" validate:"required,min=1,max=10"`
}

// ImagesEncode 编码图片
func (r *CreateFeedbackRequest) ImagesEncode() ([]byte, error) {
	return json.Marshal(r.Images)
}
