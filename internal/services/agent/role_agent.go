// Package agent 角色代理服务
package agent

import (
	"aibuddy/internal/model"
	"aibuddy/internal/repository"
	"aibuddy/pkg/helpers"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"text/template"
	"time"

	agentmodel "trpc.group/trpc-go/trpc-agent-go/model"
	"trpc.group/trpc-go/trpc-agent-go/runner"
)

// RoleAgentService 角色代理服务
type RoleAgentService struct {
	ChatDialogueRepository *repository.ChatDialogueRepository
}

// NewRoleAgentService 创建角色代理服务
func NewRoleAgentService() *RoleAgentService {
	return &RoleAgentService{
		ChatDialogueRepository: repository.NewChatDialogueRepository(),
	}
}

// BuildRoleAgentPrompt 构建角色代理提示词
func (s *RoleAgentService) BuildRoleAgentPrompt(dialogue string) (string, error) {
	tmpl, err := template.New("role_agent").Parse(RoleAgentPrompt)
	if err != nil {
		return "", fmt.Errorf("parse template failed: %w", err)
	}

	data := struct {
		Dialogue string
	}{
		Dialogue: dialogue,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template failed: %w", err)
	}

	return buf.String(), nil
}

// RoleChatAgent 角色聊天代理
// startDate 开始日期
// endDate 结束日期
// roleAgentName 角色代理名称
// 查询聊天记录获取对应角色代理的聊天记录
// 生成用户角色使用的报告
func (s *RoleAgentService) RoleChatAgent(deviceID string, startDate, endDate time.Time, roleAgentName string) (*RoleAgentReport, error) {
	return helpers.Retry(10, time.Minute*20, func() (*RoleAgentReport, error) {
		dialogues, err := s.ChatDialogueRepository.GetChatDialogue(deviceID, startDate, endDate, roleAgentName)
		if err != nil {
			return nil, err
		}
		if len(dialogues) < 2 {
			return nil, helpers.SkipRetry
		}

		content, err := s.FormatChatToTemplate(dialogues)
		if err != nil {
			return nil, err
		}

		prompt, err := s.BuildRoleAgentPrompt(content)
		if err != nil {
			return nil, err
		}

		agentModel := NewAgentModel(
			s.AgentOutputFormat(),
			"请严格按照需求输出",
		)

		r := runner.NewRunner("role_agent", agentModel)

		eventCh, err := r.Run(context.Background(), "role_agent", "role_agent", agentmodel.Message{
			Role:    agentmodel.RoleUser,
			Content: prompt,
		})

		if err != nil {
			return nil, err
		}

		var result *RoleAgentReport
		for evt := range eventCh {
			if evt.IsError() {
				return nil, fmt.Errorf("agent error: %v", evt.Error)
			}

			if evt.StructuredOutput != nil {
				if report, ok := evt.StructuredOutput.(*RoleAgentReport); ok {
					result = report
				}
			}
		}

		return result, nil
	})
}

// FormatChatToTemplate 格式化聊天记录为模板
type FormatChatToTemplate struct {
	Query      string `json:"query"`
	QueryTime  string `json:"query_time"`
	Answer     string `json:"answer"`
	AnswerTime string `json:"answer_time"`
}

// FormatChatToTemplate 格式化聊天记录为模板
func (s *RoleAgentService) FormatChatToTemplate(dialogue []*model.ChatDialogue) (string, error) {
	formatChatToTemplate := make([]*FormatChatToTemplate, 0, len(dialogue))

	for _, d := range dialogue {
		formatChatToTemplate = append(formatChatToTemplate, &FormatChatToTemplate{
			Query:      d.Question,
			QueryTime:  d.QuestionTime.Format(time.DateTime),
			Answer:     d.Answer,
			AnswerTime: d.AnswerTime.Format(time.DateTime),
		})
	}

	jsonData, err := json.Marshal(formatChatToTemplate)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// AgentOutputFormat 返回模型输出的 JSON Schema 格式定义
func (s *RoleAgentService) AgentOutputFormat() *RoleAgentReport {
	return &RoleAgentReport{}
}

// ConversationSummary 对话总结
type ConversationSummary struct {
	UpdatedAt   string              `json:"updatedAt"`
	Summary     string              `json:"summary"`
	Topics      []ConversationTopic `json:"topics"`
	FocusPoints []string            `json:"focusPoints"`
}

// Encode 编码 ConversationSummary 为 JSON
func (s *ConversationSummary) Encode() ([]byte, error) {
	return json.Marshal(s)
}

// ConversationTopic 对话话题
type ConversationTopic struct {
	Title string `json:"title"`
	Desc  string `json:"desc"`
}

// Period 时间段
type Period struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// EmotionAnalysis 情绪分析
type EmotionAnalysis struct {
	Period       Period       `json:"period"`
	SummaryDate  string       `json:"summary_date"`
	OverallTrend OverallTrend `json:"overall_trend"`
}

// Encode 编码 EmotionAnalysis 为 JSON
func (s *EmotionAnalysis) Encode() ([]byte, error) {
	return json.Marshal(s)
}

// OverallTrend 总体趋势
type OverallTrend struct {
	Trend               string               `json:"trend"`
	DominantEmotion     string               `json:"dominant_emotion"`
	AverageIntensity    int                  `json:"average_intensity"`
	KeyEmotionalMoments []KeyEmotionalMoment `json:"key_emotional_moments"`
}

// KeyEmotionalMoment 关键情绪时刻
type KeyEmotionalMoment struct {
	Date    string `json:"date"`
	Note    string `json:"note"`
	Event   string `json:"event"`
	Emotion string `json:"emotion"`
}

// RoleAgentReport 角色代理报告（顶层结构）
type RoleAgentReport struct {
	ConversationSummary *ConversationSummary `json:"conversationSummary"`
	EmotionAnalysis     *EmotionAnalysis     `json:"EmotionAnalysis"`
}
