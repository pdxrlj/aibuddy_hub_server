// Package agent 情绪预警服务
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

// WarningTriggerResult 预警触发结果
type WarningTriggerResult struct {
	TriggerWarning bool `json:"trigger_warning"`
}

// WarningResult 预警结果
type WarningResult struct {
	TriggerWarning     bool          `json:"trigger_warning"`
	WarningLevel       string        `json:"warning_level"`
	WarningTypes       []string      `json:"warning_types"`
	Confidence         float64       `json:"confidence"`
	WarningReason      WarningReason `json:"warning_reason"`
	Evidence           []Evidence    `json:"evidence"`
	ParentSuggestions  []string      `json:"parent_suggestions"`
	NeedManualFollowup bool          `json:"need_manual_followup"`
	PrivacyRisk        RiskInfo      `json:"privacy_risk"`
	ScamRisk           RiskInfo      `json:"scam_risk"`
	EmotionalRisk      EmotionalRisk `json:"emotional_risk"`
	OverallAssessment  string        `json:"overall_assessment"`
}

// WarningReason 预警原因
type WarningReason struct {
	Summary     string `json:"summary"`
	WhyPrompted string `json:"why_prompted"`
	AIAction    string `json:"ai_action"`
}

// Evidence 证据
type Evidence struct {
	Speaker  string `json:"speaker"`
	Content  string `json:"content"`
	RiskType string `json:"risk_type"`
	Severity string `json:"severity"`
}

// RiskInfo 风险信息
type RiskInfo struct {
	HasPrivacyRisk bool     `json:"has_privacy_risk"`
	PrivacyItems   []string `json:"privacy_items"`
	HasScamRisk    bool     `json:"has_scam_risk"`
	ScamSignals    []string `json:"scam_signals"`
}

// EmotionalRisk 情绪风险
type EmotionalRisk struct {
	HasEmotionalRisk bool     `json:"has_emotional_risk"`
	EmotionalSignals []string `json:"emotional_signals"`
}

// Encode 编码 WarningResult 为 JSON
func (w *WarningResult) Encode() ([]byte, error) {
	return json.Marshal(w)
}

// ToEmotion 将 WarningResult 转换为 Emotion 模型
// deviceID 设备ID
// dialogueID 对话ID
func (w *WarningResult) ToEmotion(deviceID string, dialogueID int64) (*model.Emotion, error) {
	warningTypesJSON, err := json.Marshal(w.WarningTypes)
	if err != nil {
		return nil, fmt.Errorf("marshal warning_types failed: %w", err)
	}

	warningReasonJSON, err := json.Marshal(w.WarningReason)
	if err != nil {
		return nil, fmt.Errorf("marshal warning_reason failed: %w", err)
	}

	evidenceJSON, err := json.Marshal(w.Evidence)
	if err != nil {
		return nil, fmt.Errorf("marshal evidence failed: %w", err)
	}

	suggestionsJSON, err := json.Marshal(w.ParentSuggestions)
	if err != nil {
		return nil, fmt.Errorf("marshal parent_suggestions failed: %w", err)
	}

	privacyRiskJSON, err := json.Marshal(w.PrivacyRisk)
	if err != nil {
		return nil, fmt.Errorf("marshal privacy_risk failed: %w", err)
	}

	scamRiskJSON, err := json.Marshal(w.ScamRisk)
	if err != nil {
		return nil, fmt.Errorf("marshal scam_risk failed: %w", err)
	}

	emotionalRiskJSON, err := json.Marshal(w.EmotionalRisk)
	if err != nil {
		return nil, fmt.Errorf("marshal emotional_risk failed: %w", err)
	}

	return &model.Emotion{
		DeviceID:           deviceID,
		DialogueID:         dialogueID,
		TriggerWarning:     w.TriggerWarning,
		WarningLevel:       w.WarningLevel,
		WarningTypes:       warningTypesJSON,
		Confidence:         w.Confidence,
		WarningReason:      warningReasonJSON,
		Evidence:           evidenceJSON,
		ParentSuggestions:  suggestionsJSON,
		NeedManualFollowup: w.NeedManualFollowup,
		PrivacyRisk:        privacyRiskJSON,
		ScamRisk:           scamRiskJSON,
		EmotionalRisk:      emotionalRiskJSON,
		OverallAssessment:  w.OverallAssessment,
	}, nil
}

// EmotionWarningService 情绪预警服务
type EmotionWarningService struct {
	emotionRepo *repository.EmotionRepo
}

// NewEmotionWarningService 创建情绪预警服务
func NewEmotionWarningService() *EmotionWarningService {
	return &EmotionWarningService{
		emotionRepo: repository.NewEmotionRepo(),
	}
}

// BuildWarningPrompt 构建预警提示词
func (s *EmotionWarningService) BuildWarningPrompt(chatHistory string) (string, error) {
	tmpl, err := template.New("emotion").Parse(EmotionPrompt)
	if err != nil {
		return "", fmt.Errorf("parse template failed: %w", err)
	}

	data := struct {
		ChatHistory string
	}{
		ChatHistory: chatHistory,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template failed: %w", err)
	}

	return buf.String(), nil
}

// BuildWarningTriggerPrompt 构建预警触发判断提示词
func (s *EmotionWarningService) BuildWarningTriggerPrompt(chatHistory string) (string, error) {
	tmpl, err := template.New("emotion").Parse(WarningPrompt)
	if err != nil {
		return "", fmt.Errorf("parse template failed: %w", err)
	}

	data := struct {
		ChatHistory string
	}{
		ChatHistory: chatHistory,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template failed: %w", err)
	}

	return buf.String(), nil
}

// NewWarningResult 创建空预警结果
func NewWarningResult() *WarningResult {
	return &WarningResult{
		TriggerWarning:     false,
		WarningLevel:       "",
		WarningTypes:       []string{},
		Confidence:         0.0,
		WarningReason:      WarningReason{},
		Evidence:           []Evidence{},
		ParentSuggestions:  []string{},
		NeedManualFollowup: false,
		PrivacyRisk:        RiskInfo{},
		ScamRisk:           RiskInfo{},
		EmotionalRisk:      EmotionalRisk{},
		OverallAssessment:  "",
	}
}

// ParseWarningResult 解析预警结果 JSON
func ParseWarningResult(data []byte) (*WarningResult, error) {
	result := &WarningResult{}
	err := json.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ChatDialogueFormat 聊天对话格式
type ChatDialogueFormat struct {
	Query      string `json:"query"`
	QueryTime  string `json:"query_time"`
	Answer     string `json:"answer"`
	AnswerTime string `json:"answer_time"`
}

// FormatChatToTemplate 格式化聊天记录为模板
func (s *EmotionWarningService) FormatChatToTemplate(dialogues []*model.ChatDialogue) (string, error) {
	formatted := make([]*ChatDialogueFormat, 0, len(dialogues))

	for _, d := range dialogues {
		formatted = append(formatted, &ChatDialogueFormat{
			Query:      d.Question,
			QueryTime:  d.QuestionTime.Format(time.DateTime),
			Answer:     d.Answer,
			AnswerTime: d.AnswerTime.Format(time.DateTime),
		})
	}

	jsonData, err := json.Marshal(formatted)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// AgentOutputFormat 返回模型输出的格式定义
func (s *EmotionWarningService) AgentOutputFormat() *WarningResult {
	return &WarningResult{}
}

// AgentTriggerOutputFormat 返回预警触发判断的输出格式定义
func (s *EmotionWarningService) AgentTriggerOutputFormat() *WarningTriggerResult {
	return &WarningTriggerResult{}
}

// CheckWarningTrigger 检查是否需要触发预警
// 先调用门控模型判断是否需要触发预警
func (s *EmotionWarningService) CheckWarningTrigger(chatHistory string) (*WarningTriggerResult, error) {
	prompt, err := s.BuildWarningTriggerPrompt(chatHistory)
	if err != nil {
		return nil, err
	}

	agentModel := NewAgentModel(
		s.AgentTriggerOutputFormat(),
		"请严格按照需求输出",
	)

	r := runner.NewRunner("emotion_warning_trigger", agentModel)

	eventCh, err := r.Run(context.Background(),
		"emotion_warning_trigger",
		"emotion_warning_trigger",
		agentmodel.Message{
			Role:    agentmodel.RoleUser,
			Content: prompt,
		})

	if err != nil {
		return nil, err
	}

	var result *WarningTriggerResult
	for evt := range eventCh {
		if evt.IsError() {
			return nil, fmt.Errorf("agent error: %v", evt.Error)
		}

		if evt.StructuredOutput != nil {
			if trigger, ok := evt.StructuredOutput.(*WarningTriggerResult); ok {
				result = trigger
			}
		}
	}

	return result, nil
}

// GenerateDetailedWarning 生成详细预警数据
func (s *EmotionWarningService) GenerateDetailedWarning(content string) (*WarningResult, error) {
	prompt, err := s.BuildWarningPrompt(content)
	if err != nil {
		return nil, err
	}

	agentModel := NewAgentModel(
		s.AgentOutputFormat(),
		"请严格按照需求输出",
	)

	r := runner.NewRunner("emotion_warning", agentModel)

	eventCh, err := r.Run(context.Background(), "emotion_warning", "emotion_warning", agentmodel.Message{
		Role:    agentmodel.RoleUser,
		Content: prompt,
	})
	if err != nil {
		return nil, err
	}

	var result *WarningResult
	for evt := range eventCh {
		if evt.IsError() {
			return nil, fmt.Errorf("agent error: %v", evt.Error)
		}

		if evt.StructuredOutput != nil {
			if warning, ok := evt.StructuredOutput.(*WarningResult); ok {
				result = warning
			}
		}
	}

	if result != nil {
		result.TriggerWarning = true
	}

	return result, nil
}

// GenerateWarning 生成预警结果
// dialogues 聊天记录
// 使用模型分析聊天记录，生成预警结果
func (s *EmotionWarningService) GenerateWarning(dialogues []*model.ChatDialogue) (*WarningResult, error) {
	return helpers.Retry(3, time.Minute*5, func() (*WarningResult, error) {
		if len(dialogues) < 1 {
			return nil, helpers.SkipRetry
		}

		content, err := s.FormatChatToTemplate(dialogues)
		if err != nil {
			return nil, err
		}

		triggerResult, err := s.CheckWarningTrigger(content)
		if err != nil {
			return nil, err
		}

		if triggerResult == nil || !triggerResult.TriggerWarning {
			return NewWarningResult(), nil
		}

		return s.GenerateDetailedWarning(content)
	})
}

// CreateEmotion 创建情绪
func (s *EmotionWarningService) CreateEmotion(ctx context.Context, emotion *model.Emotion) error {
	return s.emotionRepo.CreateEmotion(ctx, emotion)
}
