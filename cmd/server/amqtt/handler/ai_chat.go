package handler

import (
	ai "aibuddy/aiframe/ai_chat"
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"aibuddy/internal/services/agent"
	"aibuddy/internal/services/cache"
	"aibuddy/pkg/baidu"
	"aibuddy/pkg/config"
	"aibuddy/pkg/flash"
	"aibuddy/pkg/mqtt"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/cast"
)

// AiChatHandler AI对话处理器
type AiChatHandler struct {
	cache          flash.Flash
	EmotionService *agent.EmotionWarningService
	Cache          flash.Flash
}

// NewAiChatHandler 创建AI对话处理器
func NewAiChatHandler() *AiChatHandler {
	return &AiChatHandler{
		cache:          cache.Flash(),
		EmotionService: agent.NewEmotionWarningService(),
		Cache:          cache.Flash(),
	}
}

// Chat 处理AI对话
func (h *AiChatHandler) Chat(ctx *mqtt.Context) {
	defer ctx.Message.Ack()

	deviceID := ctx.Params["device_id"]
	var msg ai.Chat

	if err := msg.Decode(ctx.Payload); err != nil {
		slog.Error("[MQTT] Decode failed", "error", err)
		return
	}

	// 角色切换
	if msg.Type == ai.ChatTypeSwitchRole {
		h.handleSwitchRole(deviceID, msg.Role)
		return
	}

	cacheKey := fmt.Sprintf("ai_chat:%s:%s:%s", deviceID, msg.Sid, msg.Type)
	slog.Info("[MQTT] Chat request", "device_id", deviceID, "sid", msg.Sid, "type", msg.Type)
	if err := h.cache.Set(cacheKey, time.Now().Unix(), time.Hour*24); err != nil {
		slog.Error("[MQTT] Set cache failed", "error", err)
		return
	}

	if msg.Type == ai.ChatTypeEnd {
		slog.Info("[Baidu] Chat", "对话结束开始下载", deviceID)
		// 获取对话开始时间
		var startTime int64
		startKey := fmt.Sprintf("ai_chat:%s:%s:%s", deviceID, msg.Sid, ai.ChatTypeStart)
		if val, err := h.cache.Get(startKey); err == nil {
			startTime = cast.ToInt64(val)
		} else {
			// 没有开始时间，默认使用当前时间减去对话时长
			// startTime = time.Now().Add(-time.Duration(msg.Dur) * time.Second).Unix()

			return
		}

		// 下载数据并触发预警
		go func() {
			slog.Info("[Baidu] Chat", "开始准备下载", deviceID)
			if err := h.downloadDialogues(deviceID, startTime, time.Now().Unix()); err != nil {
				slog.Error("[MQTT] Download dialogues failed", "error", err)
				return
			}

			// 查询刚下载的对话记录
			dialogues, err := query.ChatDialogue.
				Where(query.ChatDialogue.DeviceID.Eq(deviceID)).
				Where(query.ChatDialogue.QuestionTime.Gte(time.Unix(startTime, 0))).
				Find()
			if err != nil {
				slog.Error("[MQTT] Query dialogues failed", "error", err)
				return
			}

			// 触发情绪预警
			if _, err := h.TriggerWarning(context.Background(), deviceID, dialogues); err != nil {
				slog.Error("[MQTT] Trigger warning failed", "error", err)
			}
			slog.Info("[MQTT] Trigger warning success", "device_id", deviceID)
		}()
	}
}

// handleSwitchRole 处理角色切换
func (h *AiChatHandler) handleSwitchRole(deviceID, role string) {
	_, err := query.Device.Where(query.Device.DeviceID.Eq(deviceID)).Update(query.Device.AgentName, role)
	if err != nil {
		slog.Error("[MQTT] Update device agent name failed", "error", err)
	}
}

// pageNo	Integer	当前页
// pageSize	Integer	当前页查询到的数量，如果返回值比输入的pageSize小，则表明当前页已是最后一页
func (h *AiChatHandler) downloadDialogues(deviceID string, beginTime, endTime int64) error {
	pageSize := 100
	pageNo := 1

	client := baidu.NewDialogues()
	var allDialogues []baidu.DialogueItem
	slog.Info("[BaiDu] downloadDialogues", "开始时间", beginTime, "结束时间", endTime)
	// 查询当前对话使用的角色信息
	deviceInfo, err := query.Device.Where(query.Device.DeviceID.Eq(deviceID)).First()
	if err != nil {
		return err
	}

	agentName := deviceInfo.AgentName

	// 分页获取所有对话记录
	for {
		dialogues, err := client.GetDialogues(&baidu.DialoguesRequest{
			AppID:     config.Instance.Baidu.AppID,
			UserID:    deviceID,
			PageNo:    pageNo,
			PageSize:  pageSize,
			BeginTime: beginTime,
			EndTime:   endTime,
		})
		if err != nil {
			slog.Info("[Baidu] downloadDialogues", "获取对话记录失败", err)
			return err
		}

		allDialogues = append(allDialogues, dialogues.Data...)

		// 如果返回数量小于请求的pageSize，说明已是最后一页
		if dialogues.PageSize < pageSize {
			break
		}
		pageNo++
	}
	slog.Info("[Baidu] downloadDialogues", "共", fmt.Sprintf("%d 条", len(allDialogues)))
	// 将对话记录配对保存（QUESTION + ANSWER）
	dialogueModels := h.pairDialogues(deviceID, agentName, allDialogues)
	if len(dialogueModels) == 0 {
		return nil
	}

	// 批量保存到数据库
	return query.ChatDialogue.CreateInBatches(dialogueModels, 100)
}

// pairDialogues 将对话记录配对（QUESTION + ANSWER）
// 假设数据格式为交替的 Q, A, Q, A...
func (h *AiChatHandler) pairDialogues(deviceID string, agentName string, items []baidu.DialogueItem) []*model.ChatDialogue {
	var result []*model.ChatDialogue

	for i := 0; i < len(items); i++ {
		if items[i].Type == "QUESTION" {
			question := items[i]
			// 紧跟的下一个应该是 ANSWER
			if i+1 < len(items) && items[i+1].Type == "ANSWER" {
				answer := items[i+1]
				result = append(result, &model.ChatDialogue{
					DeviceID:     deviceID,
					AgentName:    agentName,
					Question:     question.Text,
					QuestionTime: time.Unix(question.Timestamp, 0),
					Answer:       answer.Text,
					AnswerTime:   time.Unix(answer.Timestamp, 0),
				})
				i++ // 跳过已配对的 ANSWER
			}
		}
	}

	return result
}

// TriggerWarning 情绪预警服务
func (h *AiChatHandler) TriggerWarning(ctx context.Context, deviceID string, dialogues []*model.ChatDialogue) (*agent.WarningResult, error) {
	if len(dialogues) == 0 {
		return nil, nil
	}

	result, err := h.EmotionService.GenerateWarning(dialogues)
	if err != nil {
		slog.Error("[MQTT] Generate warning failed", "error", err)
		return nil, err
	}

	slog.Info("[MQTT] Warning result",
		"trigger_warning", result.TriggerWarning,
		"warning_level", result.WarningLevel,
		"confidence", result.Confidence,
	)

	if result.TriggerWarning {
		var dialogueID int64
		if len(dialogues) > 0 {
			dialogueID = dialogues[len(dialogues)-1].ID
		}

		emotion, err := result.ToEmotion(deviceID, dialogueID)
		if err != nil {
			slog.Error("[MQTT] Convert to emotion failed", "error", err)
			return result, nil
		}

		if err := h.EmotionService.CreateEmotion(ctx, emotion); err != nil {
			slog.Error("[MQTT] Create emotion failed", "error", err)
			return result, nil
		}

		slog.Info("[MQTT] Emotion saved", "device_id", deviceID, "warning_level", result.WarningLevel)
	}

	return result, nil
}
