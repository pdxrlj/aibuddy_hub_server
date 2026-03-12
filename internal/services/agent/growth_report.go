// Package agent 提供代理服务
package agent

import (
	"aibuddy/internal/repository"
	"aibuddy/pkg/config"
	"aibuddy/pkg/helpers"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	agentmodel "trpc.group/trpc-go/trpc-agent-go/model"
	"trpc.group/trpc-go/trpc-agent-go/runner"
)

var tracer = func() trace.Tracer {
	return otel.Tracer(config.Instance.Tracer.ServiceName)
}

// GrowthReport 群组报告
type GrowthReport struct {
	NfcRepository                *repository.NFCRepository
	EmotionRepository            *repository.EmotionRepo
	MessageRepository            *repository.DeviceMessageRepo
	PomodoroRepository           *repository.PomodoroClockRepo
	DeviceRelationshipRepository *repository.DeviceRelationshipRepo
	DialogueRepository           *repository.ChatDialogueRepository

	DeviceRepository *repository.DeviceRepo
}

// NewGroupReport 创建群组报告
func NewGroupReport() *GrowthReport {
	return &GrowthReport{
		NfcRepository:                repository.NewNFCRepository(),
		EmotionRepository:            repository.NewEmotionRepo(),
		MessageRepository:            repository.NewDeviceMessageRepo(),
		PomodoroRepository:           repository.NewPomodoroClockRepo(),
		DeviceRepository:             repository.NewDeviceRepo(),
		DeviceRelationshipRepository: repository.NewDeviceRelationshipRepo(),
		DialogueRepository:           repository.NewChatDialogueRepository(),
	}
}

// MemoryCapsuleSummaryData 记忆胶囊使用记录总结数据
type MemoryCapsuleSummaryData struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

// GetNfcUsageCount 使用次数整理
func (g *GrowthReport) GetNfcUsageCount(ctx context.Context, deviceID string, startTime, endTime time.Time) []*MemoryCapsuleSummaryData {
	_, span := tracer().Start(ctx, "GrowthReport.GetNfcUsageCount")
	defer span.End()

	nfcData, err := g.NfcRepository.GetNfcDataByDeviceID(deviceID, startTime, endTime)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		span.SetAttributes(attribute.String("start_time", startTime.Format(time.RFC3339)))
		span.SetAttributes(attribute.String("end_time", endTime.Format(time.RFC3339)))
		slog.Error("[GrowthReport] GetNfcUsageCount error", "error", err.Error())
		return nil
	}

	countMap := make(map[string]int)
	for _, nfc := range nfcData {
		countMap[nfc.Ctype]++
	}

	memoryCapsuleSummaryData := make([]*MemoryCapsuleSummaryData, 0, len(countMap))
	for t, count := range countMap {
		memoryCapsuleSummaryData = append(memoryCapsuleSummaryData, &MemoryCapsuleSummaryData{
			Type:  t,
			Count: count,
		})
	}

	return memoryCapsuleSummaryData
}

// PomodoroRecordSummaryData 番茄钟使用记录总结数据
type PomodoroRecordSummaryData struct {
	UseCount         int `json:"use_count"`
	TotalDurationMin int `json:"total_duration_min"`
	DistractionCount int `json:"distraction_count"`
}

// GetPomodoroRecordSummary 获取番茄钟使用记录
func (g *GrowthReport) GetPomodoroRecordSummary(ctx context.Context, deviceID string, startTime, endTime time.Time) *PomodoroRecordSummaryData {
	_, span := tracer().Start(ctx, "GrowthReport.GetPomodoroRecordSummary")
	defer span.End()

	pomodoroRecords, err := g.PomodoroRepository.GetPomodoroClockByDeviceID(ctx, deviceID, startTime, endTime)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		span.SetAttributes(attribute.String("start_time", startTime.Format(time.RFC3339)))
		span.SetAttributes(attribute.String("end_time", endTime.Format(time.RFC3339)))
		slog.Error("[GrowthReport] GetPomodoroRecordSummary error", "error", err.Error())
		return nil
	}

	useCount := len(pomodoroRecords)
	totalDurationMin := 0
	distractionCount := 0
	for _, pomodoroRecord := range pomodoroRecords {
		totalDurationMin += pomodoroRecord.StudyDuration
		if pomodoroRecord.DistractionDuration > 0 {
			distractionCount++
		}
	}

	return &PomodoroRecordSummaryData{
		UseCount:         useCount,
		TotalDurationMin: totalDurationMin,
		DistractionCount: distractionCount,
	}
}

// FamilyInteractionSummary 家庭互动总结
type FamilyInteractionSummary struct {
	FamilyInteractions []*FamilyInteractionSummaryData `json:"family_interactions"`
	FriendAddedCount   int                             `json:"friend_added_count"`
	FriendChatCount    int                             `json:"friend_chat_count"`
}

// FamilyInteractionSummaryData 家庭互动总结数据
type FamilyInteractionSummaryData struct {
	MemberName string `json:"member_name"`
	ChatCount  int    `json:"chat_count"`
}

// FriendAddedCountData 好友添加次数
type FriendAddedCountData struct {
	FriendAddedCount int `json:"friend_added_count"`
}

// FriendChatCountData 好友聊天次数
type FriendChatCountData struct {
	FriendChatCount int `json:"friend_chat_count"`
}

// GetInteractionSummary 获取互动总结
func (g *GrowthReport) GetInteractionSummary(ctx context.Context, deviceID string, startTime, endTime time.Time) *FamilyInteractionSummary {
	_, span := tracer().Start(ctx, "GrowthReport.GetInteractionSummary")
	defer span.End()

	devices, err := g.DeviceRelationshipRepository.GetFriendsByDeviceID(ctx, deviceID, startTime, endTime)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		span.SetAttributes(attribute.String("start_time", startTime.Format(time.RFC3339)))
		span.SetAttributes(attribute.String("end_time", endTime.Format(time.RFC3339)))
		slog.Error("[GrowthReport] GetInteractionSummary error", "error", err.Error())
		return nil
	}

	friendAddCount := len(devices)

	messageData, err := g.MessageRepository.GetMessageListByDeviceID(ctx, deviceID, startTime, endTime)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		span.SetAttributes(attribute.String("start_time", startTime.Format(time.RFC3339)))
		span.SetAttributes(attribute.String("end_time", endTime.Format(time.RFC3339)))
		slog.Error("[GrowthReport] GetInteractionSummary error", "error", err.Error())
		return nil
	}

	friendChatCount := 0

	chatCountMap := make(map[string]int)
	for _, message := range messageData {
		if strings.Contains(message.ToDeviceID, ":") || strings.Contains(message.FromDeviceID, ":") {
			if message.Device != nil && message.Device.Relation != "" {
				chatCountMap[message.Device.Relation]++
			}
		} else {
			friendChatCount++
		}
	}

	familyInteractions := make([]*FamilyInteractionSummaryData, 0, len(chatCountMap))
	for memberName, count := range chatCountMap {
		familyInteractions = append(familyInteractions, &FamilyInteractionSummaryData{
			MemberName: memberName,
			ChatCount:  count,
		})
	}

	return &FamilyInteractionSummary{
		FamilyInteractions: familyInteractions,
		FriendAddedCount:   friendAddCount,
		FriendChatCount:    friendChatCount,
	}
}

// DeviceBaseInfo 用户设备的基础信息
type DeviceBaseInfo struct {
	NickName string `json:"child_name"`
	Gender   string `json:"child_gender"`
	Age      int    `json:"child_age"`
}

// GetDeviceBaseInfo 获取用户设备的基础信息
func (g *GrowthReport) GetDeviceBaseInfo(ctx context.Context, deviceID string) *DeviceBaseInfo {
	_, span := tracer().Start(ctx, "GrowthReport.GetDeviceBaseInfo")
	defer span.End()

	device, err := g.DeviceRepository.GetDeviceInfo(ctx, deviceID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		slog.Error("[GrowthReport] GetDeviceBaseInfo error", "error", err.Error())
		return nil
	}

	if device.DeviceInfo == nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		slog.Error("[GrowthReport] DeviceInfo is nil", "device_id", deviceID)
		return nil
	}

	age := 0
	if !device.DeviceInfo.Birthday.IsZero() {
		now := time.Now()
		age = now.Year() - device.DeviceInfo.Birthday.Year()
		if now.YearDay() < device.DeviceInfo.Birthday.YearDay() {
			age--
		}
	}

	return &DeviceBaseInfo{
		NickName: device.DeviceInfo.NickName,
		Gender:   device.DeviceInfo.Gender,
		Age:      age,
	}
}

// AlertTypeData 预警类型统计
type AlertTypeData struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

// ConversationWarning 对话预警
type ConversationWarning struct {
	AlertCount int              `json:"alert_count"`
	AlertTypes []*AlertTypeData `json:"alert_types"`
}

// GetConversationWarning 获取对话预警统计
func (g *GrowthReport) GetConversationWarning(ctx context.Context, deviceID string, startTime, endTime time.Time) *ConversationWarning {
	_, span := tracer().Start(ctx, "GrowthReport.GetConversationWarning")
	defer span.End()

	emotions, err := g.EmotionRepository.GetEmotionsByDeviceID(ctx, deviceID, startTime, endTime)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		span.SetAttributes(attribute.String("start_time", startTime.Format(time.RFC3339)))
		span.SetAttributes(attribute.String("end_time", endTime.Format(time.RFC3339)))
		slog.Error("[GrowthReport] GetConversationWarning error", "error", err.Error())
		return nil
	}

	alertCount := len(emotions)
	typeCountMap := make(map[string]int)

	for _, emotion := range emotions {
		var warningTypes []string
		if len(emotion.WarningTypes) > 0 {
			if err := json.Unmarshal(emotion.WarningTypes, &warningTypes); err != nil {
				slog.Warn("[GrowthReport] unmarshal warning_types failed", "error", err.Error())
				continue
			}
		}

		for _, t := range warningTypes {
			typeCountMap[t]++
		}
	}

	alertTypes := make([]*AlertTypeData, 0, len(typeCountMap))
	for t, count := range typeCountMap {
		alertTypes = append(alertTypes, &AlertTypeData{
			Type:  t,
			Count: count,
		})
	}

	return &ConversationWarning{
		AlertCount: alertCount,
		AlertTypes: alertTypes,
	}
}

// StageOneInput 第一阶段提示词输入数据
type StageOneInput struct {
	ReportMeta   *DeviceBaseInfo           `json:"report_meta"`
	ChatLogs     []*DialoguePromptData     `json:"chat_logs"`
	FeatureUsage *FeatureUsageData         `json:"feature_usage"`
	SocialLogs   *FamilyInteractionSummary `json:"social_logs"`
	SafetyAlert  *ConversationWarning      `json:"safety_alert"`
}

// FeatureUsageData 功能使用数据
type FeatureUsageData struct {
	MemoryCapsule []*MemoryCapsuleSummaryData `json:"memory_capsule"`
	Pomodoro      *PomodoroRecordSummaryData  `json:"pomodoro"`
}

// BuildStageOnePrompt 构建第一阶段的提示词
func (g *GrowthReport) BuildStageOnePrompt(ctx context.Context, deviceID string, startTime, endTime time.Time) string {
	_, span := tracer().Start(ctx, "GrowthReport.BuildStageOnePrompt")
	defer span.End()

	// 1. 用户的设备基础信息 (report_meta)
	reportMeta := g.GetDeviceBaseInfo(ctx, deviceID)
	if reportMeta == nil {
		span.RecordError(errors.New("device base info is nil"))
		return ""
	}

	// 2. 用户的对话记录 (chat_logs)
	chatLogs := g.GetDialogueData(ctx, deviceID, startTime, endTime)

	// 3. 功能使用数据 (feature_usage)
	featureUsage := &FeatureUsageData{
		MemoryCapsule: g.GetNfcUsageCount(ctx, deviceID, startTime, endTime),
		Pomodoro:      g.GetPomodoroRecordSummary(ctx, deviceID, startTime, endTime),
	}

	// 4. 社交互动数据 (social_logs)
	socialLogs := g.GetInteractionSummary(ctx, deviceID, startTime, endTime)

	// 5. 风险事件提醒 (safety_alert)
	safetyAlert := g.GetConversationWarning(ctx, deviceID, startTime, endTime)

	input := &StageOneInput{
		ReportMeta:   reportMeta,
		ChatLogs:     chatLogs,
		FeatureUsage: featureUsage,
		SocialLogs:   socialLogs,
		SafetyAlert:  safetyAlert,
	}

	inputStr, err := json.Marshal(input)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		span.SetAttributes(attribute.String("start_time", startTime.Format(time.RFC3339)))
		span.SetAttributes(attribute.String("end_time", endTime.Format(time.RFC3339)))
		slog.Error("[GrowthReport] BuildStageOnePrompt error", "error", err.Error())
		return ""
	}

	return string(inputStr)
}

// DialoguePromptData 用户的对话记录
type DialoguePromptData struct {
	Query      string `json:"query"`
	QueryTime  string `json:"query_time"`
	Answer     string `json:"answer"`
	AnswerTime string `json:"answer_time"`
}

// GetDialogueData 获取用户的对话记录数据
func (g *GrowthReport) GetDialogueData(ctx context.Context, deviceID string, startTime, endTime time.Time) []*DialoguePromptData {
	_, span := tracer().Start(ctx, "GrowthReport.GetDialogueData")
	defer span.End()

	dialogeData, err := g.DialogueRepository.GetChatDialogue(deviceID, startTime, endTime)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("device_id", deviceID))
		span.SetAttributes(attribute.String("start_time", startTime.Format(time.RFC3339)))
		span.SetAttributes(attribute.String("end_time", endTime.Format(time.RFC3339)))
		slog.Error("[GrowthReport] GetDialogueData error", "error", err.Error())
		return nil
	}

	dialoguePromptData := make([]*DialoguePromptData, 0, len(dialogeData))
	for _, dialogue := range dialogeData {
		dialoguePromptData = append(dialoguePromptData, &DialoguePromptData{
			Query:      dialogue.Question,
			QueryTime:  dialogue.QuestionTime.Format(time.DateTime),
			Answer:     dialogue.Answer,
			AnswerTime: dialogue.AnswerTime.Format(time.DateTime),
		})
	}

	return dialoguePromptData
}

// ==================== 第一阶段输出结构体 ====================

// StageOneReport 第一阶段输出报告
type StageOneReport struct {
	FactMeta           *FactMeta           `json:"fact_meta"`
	InteractionFacts   *InteractionFacts   `json:"interaction_facts"`
	SocialFacts        *SocialFacts        `json:"social_facts"`
	MemoryCapsuleFacts *MemoryCapsuleFacts `json:"memory_capsule_facts"`
	EmotionFacts       *EmotionFacts       `json:"emotion_facts"`
	MomentCandidates   []*MomentCandidate  `json:"moment_candidates"`
	LearningFacts      *LearningFacts      `json:"learning_facts"`
	RiskFacts          *RiskFacts          `json:"risk_facts"`
	TopicFacts         *TopicFacts         `json:"topic_facts"`
	PortraitFacts      *PortraitFacts      `json:"portrait_facts"`
}

// FactMeta 报告基础信息
type FactMeta struct {
	ReportType         string `json:"report_type"`
	StartDate          string `json:"start_date"`
	EndDate            string `json:"end_date"`
	ChildName          string `json:"child_name"`
	ChildGender        string `json:"child_gender"`
	ChildAge           int    `json:"child_age"`
	SourceMessageCount int    `json:"source_message_count"`
	SourceDays         int    `json:"source_days"`
	DataStatus         string `json:"data_status"`
}

// InteractionFacts 互动事实
type InteractionFacts struct {
	TotalChatCount         int        `json:"total_chat_count"`
	TopRoles               []*TopRole `json:"top_roles"`
	LongestChatDurationMin int        `json:"longest_chat_duration_min"`
	ActiveTimeRange        string     `json:"active_time_range"`
}

// TopRole 角色统计
type TopRole struct {
	RoleName  string `json:"role_name"`
	ChatCount int    `json:"chat_count"`
}

// SocialFacts 社交事实
type SocialFacts struct {
	FamilyInteractions []*FamilyInteraction `json:"family_interactions"`
	FriendAddedCount   int                  `json:"friend_added_count"`
	FriendChatCount    int                  `json:"friend_chat_count"`
}

// FamilyInteraction 家庭互动
type FamilyInteraction struct {
	MemberName string `json:"member_name"`
	ChatCount  int    `json:"chat_count"`
}

// MemoryCapsuleFacts 记忆胶囊事实
type MemoryCapsuleFacts struct {
	Count int      `json:"count"`
	Types []string `json:"types"`
}

// EmotionFacts 情绪事实
type EmotionFacts struct {
	DailyEmotions []*DailyEmotionFact `json:"daily_emotions"`
	EmotionTags   []*EmotionTagFact   `json:"emotion_tags"`
	Summary       string              `json:"summary"`
}

// DailyEmotionFact 每日情绪事实
type DailyEmotionFact struct {
	Date           string `json:"date"`
	Score          int    `json:"score"`
	Emotion        string `json:"emotion"`
	TriggerSummary string `json:"trigger_summary"`
}

// EmotionTagFact 情绪标签事实
type EmotionTagFact struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

// MomentCandidate 代表性事件候选
type MomentCandidate struct {
	MomentType string   `json:"moment_type"`
	Title      string   `json:"title"`
	Summary    string   `json:"summary"`
	Timestamp  string   `json:"timestamp"`
	Evidence   []string `json:"evidence"`
}

// LearningFacts 学习事实
type LearningFacts struct {
	AudioSummary    *AudioSummary    `json:"audio_summary"`
	PomodoroSummary *PomodoroSummary `json:"pomodoro_summary"`
}

// AudioSummary 音频总结
type AudioSummary struct {
	ListenCount      int    `json:"listen_count"`
	TotalDurationMin int    `json:"total_duration_min"`
	FavoriteContent  string `json:"favorite_content"`
}

// PomodoroSummary 番茄钟总结
type PomodoroSummary struct {
	UseCount         int `json:"use_count"`
	TotalDurationMin int `json:"total_duration_min"`
	DistractionCount int `json:"distraction_count"`
}

// RiskFacts 风险事实
type RiskFacts struct {
	AlertCount int          `json:"alert_count"`
	AlertTypes []*AlertType `json:"alert_types"`
}

// AlertType 预警类型
type AlertType struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

// TopicFacts 话题事实
type TopicFacts struct {
	CommonTopics []string `json:"common_topics"`
}

// PortraitFacts 画像事实
type PortraitFacts struct {
	Preferences     []string `json:"preferences"`
	Dislikes        []string `json:"dislikes"`
	BehaviorSignals []string `json:"behavior_signals"`
}

// AgentOutputFormat 返回模型输出的格式
func (g *GrowthReport) AgentOutputFormat() *StageOneReport {
	return &StageOneReport{}
}

// RunStageOne 执行第一阶段分析
func (g *GrowthReport) RunStageOne(ctx context.Context, deviceID string, startTime, endTime time.Time) (*StageOneReport, error) {
	return helpers.Retry(10, time.Minute*20, func() (*StageOneReport, error) {
		_, span := tracer().Start(ctx, "GrowthReport.RunStageOne")
		defer span.End()

		// 1. 构建输入数据
		input := g.BuildStageOnePrompt(ctx, deviceID, startTime, endTime)
		if input == "" {
			return nil, helpers.SkipRetry
		}

		// 2. 创建模型
		agentModel := NewAgentModel(
			g.AgentOutputFormat(),
			"请严格按照需求输出",
		)

		// 3. 运行模型
		r := runner.NewRunner("growth_report_stage_one", agentModel)

		eventCh, err := r.Run(ctx, "growth_report_stage_one", "growth_report_stage_one", agentmodel.Message{
			Role:    agentmodel.RoleUser,
			Content: StageOnePrompt + input,
		})
		if err != nil {
			span.RecordError(err)
			return nil, err
		}

		// 4. 解析输出
		var result *StageOneReport
		for evt := range eventCh {
			if evt.IsError() {
				span.RecordError(evt.Error)
				return nil, fmt.Errorf("agent error: %v", evt.Error)
			}

			if evt.StructuredOutput != nil {
				if report, ok := evt.StructuredOutput.(*StageOneReport); ok {
					result = report
				}
			}
		}

		return result, nil
	})
}

// RunGrowthReport 执行完整的成长报告生成（第一阶段 + 第二阶段）
func (g *GrowthReport) RunGrowthReport(ctx context.Context, deviceID string, startTime, endTime time.Time) (*StageTwoReport, error) {
	_, span := tracer().Start(ctx, "GrowthReport.RunGrowthReport")
	defer span.End()

	// 第一阶段：事实抽取
	stageOneReport, err := g.RunStageOne(ctx, deviceID, startTime, endTime)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("stage one failed: %w", err)
	}

	if stageOneReport == nil {
		return nil, errors.New("stage one report is nil")
	}

	// 第二阶段：生成报告
	stageTwoReport, err := g.RunStageTwo(ctx, stageOneReport)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("stage two failed: %w", err)
	}

	return stageTwoReport, nil
}

// ==================== 第二阶段输出结构体 ====================

// StageTwoReport 第二阶段输出报告
type StageTwoReport struct {
	SummaryText          string                 `json:"summary_text"`
	StatusCards          []*StatusCard          `json:"status_cards"`
	InteractionSummary   *InteractionSummary    `json:"interaction_summary"`
	SocialSummary        *SocialSummary         `json:"social_summary"`
	MemoryCapsuleSummary *MemoryCapsuleSummary  `json:"memory_capsule_summary"`
	ChildPortrait        *ChildPortrait         `json:"child_portrait"`
	KeyMoments           []*KeyMoment           `json:"key_moments"`
	EmotionTrend         *EmotionTrend          `json:"emotion_trend"`
	AudioSummary         *AudioSummary          `json:"audio_summary"`
	PomodoroSummary      *PomodoroSummaryReport `json:"pomodoro_summary"`
	SafetyAlert          *SafetyAlertReport     `json:"safety_alert"`
	NextWeekSuggestions  []*NextWeekSuggestion  `json:"next_week_suggestions"`
	ParentScripts        []*ParentScript        `json:"parent_scripts"`
	ClosingText          string                 `json:"closing_text"`
}

// StatusCard 状态卡片
type StatusCard struct {
	Key   string `json:"key"`
	Title string `json:"title"`
	Value string `json:"value"`
	Level string `json:"level"`
}

// InteractionSummary 互动总结
type InteractionSummary struct {
	TotalChatCount         int        `json:"total_chat_count"`
	TopRoles               []*TopRole `json:"top_roles"`
	LongestChatDurationMin int        `json:"longest_chat_duration_min"`
	ActiveTimeRange        string     `json:"active_time_range"`
	Summary                string     `json:"summary"`
}

// SocialSummary 社交总结
type SocialSummary struct {
	FamilyInteractions []*FamilyInteraction `json:"family_interactions"`
	FriendAddedCount   int                  `json:"friend_added_count"`
	FriendChatCount    int                  `json:"friend_chat_count"`
	SocialConclusion   string               `json:"social_conclusion"`
}

// MemoryCapsuleSummary 记忆胶囊总结
type MemoryCapsuleSummary struct {
	Count   int    `json:"count"`
	Type    string `json:"type"`
	Summary string `json:"summary"`
}

// ChildPortrait 孩子画像
type ChildPortrait struct {
	Personality  string   `json:"personality"`
	Preferences  []string `json:"preferences"`
	Dislikes     []string `json:"dislikes"`
	ParentAdvice string   `json:"parent_advice"`
}

// KeyMoment 关键时刻
type KeyMoment struct {
	MomentType string `json:"moment_type"`
	Title      string `json:"title"`
	Summary    string `json:"summary"`
}

// EmotionTrend 情绪趋势
type EmotionTrend struct {
	Points  []*DailyEmotionFact `json:"points"`
	Summary string              `json:"summary"`
	Advice  string              `json:"advice"`
}

// PomodoroSummaryReport 番茄钟总结报告
type PomodoroSummaryReport struct {
	UseCount         int    `json:"use_count"`
	TotalDurationMin int    `json:"total_duration_min"`
	DistractionCount int    `json:"distraction_count"`
	Summary          string `json:"summary"`
}

// SafetyAlertReport 安全预警报告
type SafetyAlertReport struct {
	AlertCount int          `json:"alert_count"`
	AlertTypes []*AlertType `json:"alert_types"`
	Summary    string       `json:"summary"`
}

// NextWeekSuggestion 下周建议
type NextWeekSuggestion struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// ParentScript 家长话术
type ParentScript struct {
	Scenario string `json:"scenario"`
	Script   string `json:"script"`
}

// AgentOutputFormatTwo 返回第二阶段模型输出的格式
func (g *GrowthReport) AgentOutputFormatTwo() *StageTwoReport {
	return &StageTwoReport{}
}

// RunStageTwo 执行第二阶段分析
func (g *GrowthReport) RunStageTwo(ctx context.Context, stageOneReport *StageOneReport) (*StageTwoReport, error) {
	return helpers.Retry(10, time.Minute*20, func() (*StageTwoReport, error) {
		_, span := tracer().Start(ctx, "GrowthReport.RunStageTwo")
		defer span.End()

		if stageOneReport == nil {
			return nil, helpers.SkipRetry
		}
		factsJSON, err := json.Marshal(stageOneReport)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}

		agentModel := NewAgentModel(
			g.AgentOutputFormatTwo(),
			"请严格按照需求输出",
		)

		r := runner.NewRunner("growth_report_stage_two", agentModel)

		eventCh, err := r.Run(ctx, "growth_report_stage_two", "growth_report_stage_two", agentmodel.Message{
			Role:    agentmodel.RoleUser,
			Content: GrowthPrompt + string(factsJSON),
		})
		if err != nil {
			span.RecordError(err)
			return nil, err
		}

		var result *StageTwoReport
		for evt := range eventCh {
			if evt.IsError() {
				span.RecordError(evt.Error)
				return nil, fmt.Errorf("agent error: %v", evt.Error)
			}

			if evt.StructuredOutput != nil {
				if report, ok := evt.StructuredOutput.(*StageTwoReport); ok {
					result = report
				}
			}
		}

		return result, nil
	})
}
