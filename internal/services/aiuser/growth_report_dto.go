// Package aiuser 成长报告DTO
package aiuser

import (
	"aibuddy/internal/model"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// GrowthReportResponse 成长报告响应结构
type GrowthReportResponse struct {
	StartDate      string            `json:"start_date"`
	EndDate        string            `json:"end_date"`
	ChildName      string            `json:"child_name"`
	ChildGender    string            `json:"child_gender"`
	ChildAge       int               `json:"child_age"`
	DisplayContent DisplayContentDTO `json:"display_content"`
}

// DisplayContentDTO 展示内容
type DisplayContentDTO struct {
	CoverCard                  CoverCardDTO                  `json:"cover_card"`
	StatusCardsSection         StatusCardsSectionDTO         `json:"status_cards_section"`
	InteractionSummarySection  InteractionSummarySectionDTO  `json:"interaction_summary_section"`
	SocialSummarySection       json.RawMessage               `json:"social_summary_section"`
	MemoryCapsuleSection       json.RawMessage               `json:"memory_capsule_summary_section"`
	ChildPortraitSection       json.RawMessage               `json:"child_portrait_section"`
	KeyMomentsSection          KeyMomentsSectionDTO          `json:"key_moments_section"`
	EmotionTrendSection        json.RawMessage               `json:"emotion_trend_section"`
	LearningInterestSection    LearningInterestSectionDTO    `json:"learning_interest_section"`
	SafetyAlertSection         json.RawMessage               `json:"safety_alert_section"`
	NextWeekSuggestionsSection NextWeekSuggestionsSectionDTO `json:"next_week_suggestions_section"`
	ParentScriptsSection       ParentScriptsSectionDTO       `json:"parent_scripts_section"`
	ClosingSection             ClosingSectionDTO             `json:"closing_section"`
}

// CoverCardDTO 封面卡片
type CoverCardDTO struct {
	DateText    string `json:"date_text"`
	SummaryText string `json:"summary_text"`
}

// StatusCardsSectionDTO 状态卡片部分
type StatusCardsSectionDTO struct {
	Cards []json.RawMessage `json:"cards"`
}

// InteractionSummarySectionDTO 互动小结部分
type InteractionSummarySectionDTO struct {
	Title       string          `json:"title"`
	MainContent json.RawMessage `json:"main_content"`
}

// KeyMomentsSectionDTO 关键时刻部分
type KeyMomentsSectionDTO struct {
	Title   string            `json:"title"`
	Moments []json.RawMessage `json:"moments"`
}

// LearningInterestSectionDTO 学习兴趣部分
type LearningInterestSectionDTO struct {
	Title           string          `json:"title"`
	AudioSummary    json.RawMessage `json:"audio_summary"`
	PomodoroSummary json.RawMessage `json:"pomodoro_summary"`
}

// NextWeekSuggestionsSectionDTO 下周建议部分
type NextWeekSuggestionsSectionDTO struct {
	Title       string            `json:"title"`
	Suggestions []json.RawMessage `json:"suggestions"`
}

// ParentScriptsSectionDTO 家长话术部分
type ParentScriptsSectionDTO struct {
	Title   string            `json:"title"`
	Scripts []json.RawMessage `json:"scripts"`
}

// ClosingSectionDTO 结尾部分
type ClosingSectionDTO struct {
	Title       string `json:"title"`
	ClosingText string `json:"closing_text"`
}

// FormatGrowthReport 格式化成长报告为响应结构
func (s *Service) FormatGrowthReport(ctx context.Context, report *model.GrowthReport) (*GrowthReportResponse, error) {
	deviceID := report.DeviceID
	device, err := s.deviceService.GetDeviceInfo(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	// 计算年龄
	age := 0
	if device.DeviceInfo != nil && !device.DeviceInfo.Birthday.IsZero() {
		now := time.Now()
		age = now.Year() - device.DeviceInfo.Birthday.Year()
		if now.YearDay() < device.DeviceInfo.Birthday.YearDay() {
			age--
		}
	}

	// 构建响应
	result := &GrowthReportResponse{
		StartDate:   report.StartTime.Format("2006-01-02"),
		EndDate:     report.EndTime.Format("2006-01-02"),
		ChildName:   device.DeviceInfo.NickName,
		ChildGender: device.DeviceInfo.Gender,
		ChildAge:    age,
		DisplayContent: DisplayContentDTO{
			CoverCard: CoverCardDTO{
				DateText:    fmt.Sprintf("%s - %s", report.StartTime.Format("2006.01.02"), report.EndTime.Format("2006.01.02")),
				SummaryText: report.SummaryText,
			},
			StatusCardsSection: StatusCardsSectionDTO{
				Cards: toRawMessages(report.StatusCards),
			},
			InteractionSummarySection: InteractionSummarySectionDTO{
				Title:       "本周互动小结",
				MainContent: json.RawMessage(report.InteractionSummary),
			},
			SocialSummarySection: json.RawMessage(report.SocialSummary),
			MemoryCapsuleSection: json.RawMessage(report.MemoryCapsuleSummary),
			ChildPortraitSection: json.RawMessage(report.ChildPortrait),
			KeyMomentsSection: KeyMomentsSectionDTO{
				Title:   "本周三个代表时刻",
				Moments: toRawMessages(report.KeyMoments),
			},
			EmotionTrendSection: json.RawMessage(report.EmotionTrend),
			LearningInterestSection: LearningInterestSectionDTO{
				Title:           "学习 / 兴趣小收获",
				AudioSummary:    json.RawMessage(report.AudioSummary),
				PomodoroSummary: json.RawMessage(report.PomodoroSummary),
			},
			SafetyAlertSection: json.RawMessage(report.SafetyAlert),
			NextWeekSuggestionsSection: NextWeekSuggestionsSectionDTO{
				Title:       "下周可以这样陪伴孩子",
				Suggestions: toRawMessages(report.NextWeekSuggestions),
			},
			ParentScriptsSection: ParentScriptsSectionDTO{
				Title:   "家长可以这样说",
				Scripts: toRawMessages(report.ParentScripts),
			},
			ClosingSection: ClosingSectionDTO{
				Title:       "这一周，孩子的成长",
				ClosingText: report.ClosingText,
			},
		},
	}

	return result, nil
}

// toRawMessages 将 JSON 数组转换为 []json.RawMessage
func toRawMessages(data []byte) []json.RawMessage {
	if len(data) == 0 {
		return nil
	}
	var result []json.RawMessage
	_ = json.Unmarshal(data, &result)
	return result
}
