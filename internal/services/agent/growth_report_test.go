package agent

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGrowthReport_RunStageOne 测试第一阶段：事实抽取
func TestGrowthReport_RunStageOne(t *testing.T) {
	report := NewGroupReport()

	deviceID := "30:ED:A0:E9:F3:22"
	startTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	endTime := time.Date(2026, 3, 28, 23, 59, 59, 0, time.Local)

	ctx := context.Background()
	result, err := report.RunStageOne(ctx, deviceID, startTime, endTime)

	require.NoError(t, err)
	require.NotNil(t, result)

	// 打印结果
	jsonData, err := json.MarshalIndent(result, "", "  ")
	require.NoError(t, err)
	t.Logf("StageOne Result:\n%s", string(jsonData))

	// 验证关键字段
	assert.NotNil(t, result.FactMeta, "FactMeta should not be nil")
	assert.NotNil(t, result.InteractionFacts, "InteractionFacts should not be nil")
	assert.NotNil(t, result.SocialFacts, "SocialFacts should not be nil")
	assert.NotNil(t, result.EmotionFacts, "EmotionFacts should not be nil")
	assert.NotNil(t, result.LearningFacts, "LearningFacts should not be nil")
	assert.NotNil(t, result.RiskFacts, "RiskFacts should not be nil")
	assert.NotNil(t, result.TopicFacts, "TopicFacts should not be nil")
	assert.NotNil(t, result.PortraitFacts, "PortraitFacts should not be nil")

	// 验证新增字段
	if result.InteractionFacts != nil {
		t.Logf("InteractionFacts - DataCompleteness: %s, Confidence: %s",
			result.InteractionFacts.DataCompleteness,
			result.InteractionFacts.Confidence)
	}
}

// TestGrowthReport_RunStageTwo 测试第二阶段：生成报告
func TestGrowthReport_RunStageTwo(t *testing.T) {
	report := NewGroupReport()

	deviceID := "30:ED:A0:E9:F3:13"
	startTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	endTime := time.Date(2026, 3, 28, 23, 59, 59, 0, time.Local)

	ctx := context.Background()

	// 先执行第一阶段
	stageOneReport, err := report.RunStageOne(ctx, deviceID, startTime, endTime)
	require.NoError(t, err)
	require.NotNil(t, stageOneReport)

	// 执行第二阶段
	result, err := report.RunStageTwo(ctx, stageOneReport)
	require.NoError(t, err)
	require.NotNil(t, result)

	// 打印结果
	jsonData, err := json.MarshalIndent(result, "", "  ")
	require.NoError(t, err)
	t.Logf("StageTwo Result:\n%s", string(jsonData))

	// 验证关键字段
	assert.NotEmpty(t, result.SummaryText, "SummaryText should not be empty")
	assert.NotEmpty(t, result.StatusCards, "StatusCards should not be empty")
	assert.NotNil(t, result.InteractionSummary, "InteractionSummary should not be nil")
	assert.NotNil(t, result.SocialSummary, "SocialSummary should not be nil")
	assert.NotNil(t, result.ChildPortrait, "ChildPortrait should not be nil")
	assert.NotNil(t, result.EmotionTrend, "EmotionTrend should not be nil")
	assert.NotEmpty(t, result.ClosingText, "ClosingText should not be empty")
}

// TestGrowthReport_RunGrowthReport 测试完整的成长报告生成流程
func TestGrowthReport_RunGrowthReport(t *testing.T) {
	report := NewGroupReport()

	deviceID := "30:ED:A0:E9:F3:13"
	startTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	endTime := time.Date(2026, 3, 28, 23, 59, 59, 0, time.Local)

	ctx := context.Background()
	result, err := report.RunGrowthReport(ctx, deviceID, startTime, endTime)

	require.NoError(t, err)
	require.NotNil(t, result)

	// 打印完整报告
	jsonData, err := json.MarshalIndent(result, "", "  ")
	require.NoError(t, err)
	t.Logf("Full Growth Report:\n%s", string(jsonData))

	// 验证报告结构完整性
	assert.NotEmpty(t, result.SummaryText, "SummaryText should not be empty")
	assert.GreaterOrEqual(t, len(result.StatusCards), 2, "Should have at least 2 status cards")
	assert.NotNil(t, result.InteractionSummary)
	assert.NotNil(t, result.SocialSummary)
	assert.NotNil(t, result.MemoryCapsuleSummary)
	assert.NotNil(t, result.ChildPortrait)
	assert.NotNil(t, result.EmotionTrend)
	assert.NotNil(t, result.AudioSummary)
	assert.NotNil(t, result.PomodoroSummary)
	assert.NotNil(t, result.SafetyAlert)
	assert.Len(t, result.NextWeekSuggestions, 2, "Should have exactly 2 suggestions")
	assert.NotEmpty(t, result.ClosingText, "ClosingText should not be empty")

	// 验证 status_cards 的 level 字段
	for _, card := range result.StatusCards {
		assert.Contains(t, []string{"good", "normal", "weak"}, card.Level,
			"StatusCard level should be one of: good, normal, weak")
	}
}

// TestGrowthReport_BuildStageOnePrompt 测试构建第一阶段提示词
func TestGrowthReport_BuildStageOnePrompt(t *testing.T) {
	report := NewGroupReport()

	deviceID := "30:ED:A0:E9:F3:13"
	startTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	endTime := time.Date(2026, 3, 28, 23, 59, 59, 0, time.Local)

	ctx := context.Background()
	input := report.BuildStageOnePrompt(ctx, deviceID, startTime, endTime)

	assert.NotEmpty(t, input, "Input should not be empty")

	// 验证输入是否为有效 JSON
	var inputData map[string]interface{}
	err := json.Unmarshal([]byte(input), &inputData)
	assert.NoError(t, err, "Input should be valid JSON")

	// 验证关键字段存在
	assert.Contains(t, inputData, "report_meta")
	assert.Contains(t, inputData, "chat_logs")
	assert.Contains(t, inputData, "feature_usage")
	assert.Contains(t, inputData, "social_logs")
	assert.Contains(t, inputData, "safety_alert")

	t.Logf("Input JSON preview: %s", input[:min(500, len(input))])
}

// TestGrowthReport_DataCompletenessConfidence 测试数据完备度和置信度字段
func TestGrowthReport_DataCompletenessConfidence(t *testing.T) {
	report := NewGroupReport()

	deviceID := "30:ED:A0:E9:F3:13"
	startTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	endTime := time.Date(2026, 3, 28, 23, 59, 59, 0, time.Local)

	ctx := context.Background()
	result, err := report.RunStageOne(ctx, deviceID, startTime, endTime)

	require.NoError(t, err)
	require.NotNil(t, result)

	// 验证所有模块都有 data_completeness 和 confidence 字段
	validCompleteness := map[string]bool{"complete": true, "partial": true, "sparse": true}
	validConfidence := map[string]bool{"high": true, "medium": true, "low": true}

	// InteractionFacts
	if result.InteractionFacts != nil {
		assert.True(t, validCompleteness[result.InteractionFacts.DataCompleteness],
			"InteractionFacts.DataCompleteness should be valid")
		assert.True(t, validConfidence[result.InteractionFacts.Confidence],
			"InteractionFacts.Confidence should be valid")
	}

	// SocialFacts
	if result.SocialFacts != nil {
		assert.True(t, validCompleteness[result.SocialFacts.DataCompleteness],
			"SocialFacts.DataCompleteness should be valid")
		assert.True(t, validConfidence[result.SocialFacts.Confidence],
			"SocialFacts.Confidence should be valid")
	}

	// MemoryCapsuleFacts
	if result.MemoryCapsuleFacts != nil {
		assert.True(t, validCompleteness[result.MemoryCapsuleFacts.DataCompleteness],
			"MemoryCapsuleFacts.DataCompleteness should be valid")
		assert.True(t, validConfidence[result.MemoryCapsuleFacts.Confidence],
			"MemoryCapsuleFacts.Confidence should be valid")
	}

	// EmotionFacts
	if result.EmotionFacts != nil {
		assert.True(t, validCompleteness[result.EmotionFacts.DataCompleteness],
			"EmotionFacts.DataCompleteness should be valid")
		assert.True(t, validConfidence[result.EmotionFacts.Confidence],
			"EmotionFacts.Confidence should be valid")
	}

	// LearningFacts
	if result.LearningFacts != nil {
		assert.True(t, validCompleteness[result.LearningFacts.DataCompleteness],
			"LearningFacts.DataCompleteness should be valid")
		assert.True(t, validConfidence[result.LearningFacts.Confidence],
			"LearningFacts.Confidence should be valid")
	}

	// RiskFacts
	if result.RiskFacts != nil {
		assert.True(t, validCompleteness[result.RiskFacts.DataCompleteness],
			"RiskFacts.DataCompleteness should be valid")
		assert.True(t, validConfidence[result.RiskFacts.Confidence],
			"RiskFacts.Confidence should be valid")
	}

	// TopicFacts
	if result.TopicFacts != nil {
		assert.True(t, validCompleteness[result.TopicFacts.DataCompleteness],
			"TopicFacts.DataCompleteness should be valid")
		assert.True(t, validConfidence[result.TopicFacts.Confidence],
			"TopicFacts.Confidence should be valid")
	}

	// PortraitFacts
	if result.PortraitFacts != nil {
		assert.True(t, validCompleteness[result.PortraitFacts.DataCompleteness],
			"PortraitFacts.DataCompleteness should be valid")
		assert.True(t, validConfidence[result.PortraitFacts.Confidence],
			"PortraitFacts.Confidence should be valid")
	}

	// MomentCandidates
	for i, moment := range result.MomentCandidates {
		assert.True(t, validCompleteness[moment.DataCompleteness],
			"MomentCandidates[%d].DataCompleteness should be valid", i)
		assert.True(t, validConfidence[moment.Confidence],
			"MomentCandidates[%d].Confidence should be valid", i)
	}
}
