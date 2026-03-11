// Package agent 提供代理服务
package agent

import (
	"aibuddy/internal/repository"
	"aibuddy/pkg/config"
	"context"
	"log/slog"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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
