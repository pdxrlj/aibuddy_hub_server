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

// ConversationAnalysis 对话分析
type ConversationAnalysis struct {
	SummaryDate string                 `json:"summary_date"`
	Period      Period                 `json:"period"`
	Topics      Topics                 `json:"topics"`
	Interaction InteractionPreferences `json:"interaction_preferences"`
}

// Encode 编码 ConversationAnalysis 为 JSON
func (s *ConversationAnalysis) Encode() ([]byte, error) {
	return json.Marshal(s)
}

// Period 时间段
type Period struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// Topics 话题
type Topics struct {
	MainTopics           []MainTopic `json:"main_topics"`
	KeyEvents            []KeyEvent  `json:"key_events"`
	ConversationProgress string      `json:"conversation_progress"`
}

// MainTopic 主要话题
type MainTopic struct {
	Topic       string `json:"topic"`
	Frequency   string `json:"frequency"`
	Description string `json:"description"`
}

// KeyEvent 关键事件
type KeyEvent struct {
	Event          string `json:"event"`
	MentionedCount int    `json:"mentioned_count"`
	ChildFeeling   string `json:"child_feeling"`
}

// InteractionPreferences 互动偏好
type InteractionPreferences struct {
	Likes       []string    `json:"likes"`
	Dislikes    []string    `json:"dislikes"`
	ChangeTrend ChangeTrend `json:"change_trend"`
}

// ChangeTrend 变化趋势
type ChangeTrend struct {
	Trend       string `json:"trend"`
	Description string `json:"description"`
}

// EmotionAnalysis 情绪分析
type EmotionAnalysis struct {
	SummaryDate               string                    `json:"summary_date"`
	Period                    Period                    `json:"period"`
	DailyEmotions             []DailyEmotion            `json:"daily_emotions"`
	OverallTrend              OverallTrend              `json:"overall_trend"`
	FamilyCommunicationAdvice FamilyCommunicationAdvice `json:"family_communication_advice"`
}

// Encode 编码 EmotionAnalysis 为 JSON
func (s *EmotionAnalysis) Encode() ([]byte, error) {
	return json.Marshal(s)
}

// DailyEmotion 每日情绪
type DailyEmotion struct {
	Date           string       `json:"date"`
	PrimaryEmotion string       `json:"primary_emotion"`
	Intensity      int          `json:"intensity"`
	Tags           []EmotionTag `json:"tags"`
}

// EmotionTag 情绪标签
type EmotionTag struct {
	Emotion string `json:"emotion"`
	Event   string `json:"event"`
}

// OverallTrend 总体趋势
type OverallTrend struct {
	DominantEmotion     string               `json:"dominant_emotion"`
	AverageIntensity    float64              `json:"average_intensity"`
	Trend               string               `json:"trend"`
	KeyEmotionalMoments []KeyEmotionalMoment `json:"key_emotional_moments"`
}

// KeyEmotionalMoment 关键情绪时刻
type KeyEmotionalMoment struct {
	Date    string `json:"date"`
	Emotion string `json:"emotion"`
	Event   string `json:"event"`
	Note    string `json:"note"`
}

// Encode 编码 KeyEmotionalMoment 为 JSON
func (s *KeyEmotionalMoment) Encode() ([]byte, error) {
	return json.Marshal(s)
}

// FamilyCommunicationAdvice 家庭沟通建议
type FamilyCommunicationAdvice struct {
	Tips            []string `json:"tips"`
	AttentionPoints []string `json:"attention_points"`
	SuggestedTopics []string `json:"suggested_topics"`
}

// RoleAgentReport 角色代理报告（顶层结构）
type RoleAgentReport struct {
	ConversationAnalysis *ConversationAnalysis `json:"conversation_analysis"`
	EmotionAnalysis      *EmotionAnalysis      `json:"emotion_analysis"`
}

// Encode 编码 RoleAgentReport 为 JSON
func (s *RoleAgentReport) Encode() ([]byte, error) {
	return json.Marshal(s)
}
