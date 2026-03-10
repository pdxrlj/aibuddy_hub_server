// Package agent 提供代理服务
package agent

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"aibuddy/internal/repository"
	"fmt"
	"sort"
	"strings"
	"time"
)

// GrowthReport 群组报告
type GrowthReport struct {
	nfcRepository *repository.NFCRepository
}

// NewGroupReport 创建群组报告
func NewGroupReport() *GrowthReport {
	return &GrowthReport{
		nfcRepository: repository.NewNFCRepository(),
	}
}

// InteractionSummaryData 角色互动小结数据
type InteractionSummaryData struct {
	TotalChatCount int `json:"total_chat_count"` // 聊天总数
	TopRoles       []struct {
		RoleID    string `json:"role_id"`    // 角色ID
		RoleName  string `json:"role_name"`  // 角色名称
		ChatCount int    `json:"chat_count"` // 聊天次数
	} `json:"top_roles"`
	LongestChatDurationMin int    `json:"longest_chat_duration_min"` // 聊的最久的一次
	ActiveTimeRange        string `json:"active_time_range"`         // 一天中经常聊天的时间段
}

// GetInteractionSummary 获取角色互动小结
func (g *GrowthReport) GetInteractionSummary(deviceID string, startTime, endTime time.Time) *InteractionSummaryData {
	dialogues, err := query.ChatDialogue.
		Where(query.ChatDialogue.DeviceID.Eq(deviceID)).
		Where(query.ChatDialogue.CreatedAt.Between(startTime, endTime)).
		Find()
	if err != nil {
		return nil
	}

	roleCounts, hourCounts, maxDuration := g.analyzeDialogues(dialogues)
	topRoles := g.buildTopRoles(roleCounts)
	activeTimeRange := g.getActiveTimeRange(hourCounts)

	return &InteractionSummaryData{
		TotalChatCount:         len(dialogues),
		TopRoles:               topRoles,
		LongestChatDurationMin: maxDuration,
		ActiveTimeRange:        activeTimeRange,
	}
}

// analyzeDialogues 分析对话数据
func (g *GrowthReport) analyzeDialogues(dialogues []*model.ChatDialogue) (roleCounts map[string]int, hourCounts map[int]int, maxDuration int) {
	roleCounts = make(map[string]int)
	hourCounts = make(map[int]int)

	for _, d := range dialogues {
		if d.AgentName != "" {
			roleCounts[d.AgentName]++
		}

		duration := int(d.AnswerTime.Sub(d.QuestionTime).Minutes())
		if duration > maxDuration && duration > 0 {
			maxDuration = duration
		}

		hourCounts[d.QuestionTime.Hour()]++
	}
	return roleCounts, hourCounts, maxDuration
}

// buildTopRoles 构建热门角色列表
func (g *GrowthReport) buildTopRoles(roleCounts map[string]int) []struct {
	RoleID    string `json:"role_id"`
	RoleName  string `json:"role_name"`
	ChatCount int    `json:"chat_count"`
} {
	topRoles := make([]struct {
		RoleID    string `json:"role_id"`
		RoleName  string `json:"role_name"`
		ChatCount int    `json:"chat_count"`
	}, 0, len(roleCounts))

	for name, count := range roleCounts {
		topRoles = append(topRoles, struct {
			RoleID    string `json:"role_id"`
			RoleName  string `json:"role_name"`
			ChatCount int    `json:"chat_count"`
		}{
			RoleID:    name,
			RoleName:  name,
			ChatCount: count,
		})
	}

	sort.Slice(topRoles, func(i, j int) bool {
		return topRoles[i].ChatCount > topRoles[j].ChatCount
	})

	if len(topRoles) > 5 {
		topRoles = topRoles[:5]
	}
	return topRoles
}

// getActiveTimeRange 获取活跃时间段
func (g *GrowthReport) getActiveTimeRange(hourCounts map[int]int) string {
	if len(hourCounts) == 0 {
		return ""
	}

	activeHour, maxCount := 0, 0
	for hour, count := range hourCounts {
		if count > maxCount {
			maxCount = count
			activeHour = hour
		}
	}
	return fmt.Sprintf("%d:00-%d:00", activeHour, activeHour+1)
}

// SocialSummaryData 社交互动总结数据
type SocialSummaryData struct {
	FamilyInteractions []struct {
		MemberName string `json:"member_name"`
		ChatCount  int    `json:"chat_count"`
	} `json:"family_interactions"`
	FriendAddedCount int `json:"friend_added_count"`
	FriendChatCount  int `json:"friend_chat_count"`
}

// GetSocialSummary 获取社交互动总结
func (g *GrowthReport) GetSocialSummary(deviceID string, startTime, endTime time.Time) *SocialSummaryData {
	friendDeviceIDs, reverseDeviceIDs := g.getFriendDeviceIDs(deviceID)

	friendAddedCount := g.countNewFriends(deviceID, reverseDeviceIDs, startTime, endTime)

	messages, err := query.DeviceMessage.
		Where(query.DeviceMessage.FromDeviceID.Eq(deviceID)).
		Or(query.DeviceMessage.ToDeviceID.Eq(deviceID)).
		Where(query.DeviceMessage.CreatedAt.Between(startTime, endTime)).
		Find()
	if err != nil {
		return nil
	}

	familyChatCounts, friendChatCount, memberRelations := g.analyzeMessages(messages, deviceID, friendDeviceIDs)

	userNames := g.getUserRelationNames(memberRelations)

	return g.buildSocialSummaryData(familyChatCounts, friendChatCount, friendAddedCount, userNames)
}

// MemoryCapsuleSummaryData 记忆胶囊总结数据
type MemoryCapsuleSummaryData struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

// GetMemoryCapsuleSummary 获取记忆胶囊总结
func (g *GrowthReport) GetMemoryCapsuleSummary(deviceID string, startTime, endTime time.Time) []*MemoryCapsuleSummaryData {
	nfcData, err := g.nfcRepository.GetNfcData(deviceID, startTime, endTime)
	if err != nil {
		return nil
	}

	nfcDataMap := make(map[string]int)
	for _, nfc := range nfcData {
		nfcDataMap[nfc.Ctype]++
	}
	memoryCapsuleSummaryData := make([]*MemoryCapsuleSummaryData, 0)
	for ctype, count := range nfcDataMap {
		memoryCapsuleSummaryData = append(memoryCapsuleSummaryData, &MemoryCapsuleSummaryData{
			Type:  ctype,
			Count: count,
		})
	}
	return memoryCapsuleSummaryData
}

// getFriendDeviceIDs 获取好友设备ID列表（双向验证）
func (g *GrowthReport) getFriendDeviceIDs(deviceID string) (friendDeviceIDs map[string]bool, reverseDeviceIDs []string) {
	err := query.DeviceRelationship.
		Where(query.DeviceRelationship.TargetDeviceID.Eq(deviceID)).
		Where(query.DeviceRelationship.Status.Eq(model.RelationshipStatusAccepted.String())).
		Pluck(query.DeviceRelationship.DeviceID, &reverseDeviceIDs)
	if err != nil {
		return nil, nil
	}

	friendDeviceIDs = make(map[string]bool)
	if len(reverseDeviceIDs) > 0 {
		relationships, err := query.DeviceRelationship.
			Where(query.DeviceRelationship.DeviceID.Eq(deviceID)).
			Where(query.DeviceRelationship.Status.Eq(model.RelationshipStatusAccepted.String())).
			Where(query.DeviceRelationship.TargetDeviceID.In(reverseDeviceIDs...)).
			Find()
		if err == nil {
			for _, r := range relationships {
				friendDeviceIDs[r.TargetDeviceID] = true
			}
		}
	}
	return friendDeviceIDs, reverseDeviceIDs
}

// countNewFriends 统计新增好友数量
func (g *GrowthReport) countNewFriends(deviceID string, reverseDeviceIDs []string, startTime, endTime time.Time) int {
	if len(reverseDeviceIDs) == 0 {
		return 0
	}

	outgoingFriendships, err := query.DeviceRelationship.
		Where(query.DeviceRelationship.DeviceID.Eq(deviceID)).
		Where(query.DeviceRelationship.Status.Eq(model.RelationshipStatusAccepted.String())).
		Where(query.DeviceRelationship.TargetDeviceID.In(reverseDeviceIDs...)).
		Find()
	if err != nil {
		return 0
	}

	return g.countEstablishedFriendships(outgoingFriendships, deviceID, startTime, endTime)
}

// countEstablishedFriendships 统计已建立的好友关系数量
func (g *GrowthReport) countEstablishedFriendships(outgoingFriendships []*model.DeviceRelationship, deviceID string, startTime, endTime time.Time) int {
	var count int
	for _, outgoing := range outgoingFriendships {
		if g.isFriendshipEstablished(outgoing, deviceID, startTime, endTime) {
			count++
		}
	}
	return count
}

// isFriendshipEstablished 检查好友关系是否在指定时间范围内建立
func (g *GrowthReport) isFriendshipEstablished(outgoing *model.DeviceRelationship, deviceID string, startTime, endTime time.Time) bool {
	incoming, err := query.DeviceRelationship.
		Where(query.DeviceRelationship.DeviceID.Eq(outgoing.TargetDeviceID)).
		Where(query.DeviceRelationship.TargetDeviceID.Eq(deviceID)).
		Where(query.DeviceRelationship.Status.Eq(model.RelationshipStatusAccepted.String())).
		First()
	if err != nil || incoming == nil {
		return false
	}

	establishTime := outgoing.UpdatedAt
	if incoming.UpdatedAt.After(establishTime) {
		establishTime = incoming.UpdatedAt
	}

	return (establishTime.Equal(startTime) || establishTime.After(startTime)) &&
		(establishTime.Equal(endTime) || establishTime.Before(endTime))
}

// analyzeMessages 分析消息，统计家庭成员互动和好友聊天
func (g *GrowthReport) analyzeMessages(messages []*model.DeviceMessage, deviceID string, friendDeviceIDs map[string]bool) (
	familyChatCounts map[string]int, friendChatCount int, memberRelations map[string]map[string]string,
) {
	familyChatCounts = make(map[string]int)
	memberRelations = make(map[string]map[string]string)

	memberRelations["family"] = make(map[string]string)
	memberRelations["friend"] = make(map[string]string)

	for _, msg := range messages {
		otherID := msg.FromDeviceID
		if msg.FromDeviceID == deviceID {
			otherID = msg.ToDeviceID
		}

		if !strings.Contains(otherID, ":") {
			familyChatCounts[otherID]++
			memberRelations["family"][otherID] = otherID
		} else if friendDeviceIDs[otherID] {
			friendChatCount++
			memberRelations["friend"][otherID] = otherID
		}
	}
	return familyChatCounts, friendChatCount, memberRelations
}

// getUserRelationNames 获取用户昵称
// 家长：查询 Device 表，通过 UID 匹配获取 Relation
// 朋友：查询 DeviceInfo 表，获取 NickName
func (g *GrowthReport) getUserRelationNames(userIDs map[string]map[string]string) map[string]string {
	userNames := make(map[string]string)
	if len(userIDs) == 0 {
		return userNames
	}

	g.fillFamilyNames(userNames, userIDs["family"])
	g.fillFriendNames(userNames, userIDs["friend"])

	return userNames
}

// fillFamilyNames 填充家庭成员名称（从 Device.Relation 字段读取）
func (g *GrowthReport) fillFamilyNames(userNames map[string]string, familyIDs map[string]string) {
	if len(familyIDs) == 0 {
		return
	}

	userIDList := make([]string, 0, len(familyIDs))
	for uid := range familyIDs {
		userIDList = append(userIDList, uid)
	}

	users, err := query.User.
		Where(query.User.ID.In(stringToInt64Slice(userIDList)...)).
		Or(query.User.Phone.In(userIDList...)).
		Find()
	if err != nil || len(users) == 0 {
		return
	}

	uidList := make([]int64, 0, len(users))
	for _, u := range users {
		uidList = append(uidList, u.ID)
	}

	devices, err := query.Device.Where(query.Device.UID.In(uidList...)).Find()
	if err != nil {
		return
	}

	uidToRelation := make(map[int64]string, len(devices))
	for _, d := range devices {
		uidToRelation[d.UID] = d.Relation
	}

	for _, u := range users {
		if relation, ok := uidToRelation[u.ID]; ok {
			userNames[fmt.Sprintf("%d", u.ID)] = relation
			userNames[u.Phone] = relation
		}
	}
}

// fillFriendNames 填充朋友名称（从 DeviceInfo.NickName 字段读取）
func (g *GrowthReport) fillFriendNames(userNames map[string]string, friendIDs map[string]string) {
	if len(friendIDs) == 0 {
		return
	}

	deviceIDList := make([]string, 0, len(friendIDs))
	for deviceID := range friendIDs {
		deviceIDList = append(deviceIDList, deviceID)
	}

	deviceInfos, err := query.DeviceInfo.
		Where(query.DeviceInfo.DeviceID.In(deviceIDList...)).
		Find()
	if err != nil {
		return
	}

	for _, di := range deviceInfos {
		userNames[di.DeviceID] = di.NickName
	}
}

// buildSocialSummaryData 构建社交总结数据
func (g *GrowthReport) buildSocialSummaryData(
	familyChatCounts map[string]int,
	friendChatCount, friendAddedCount int,
	userNames map[string]string,
) *SocialSummaryData {
	familyInteractions := make([]struct {
		MemberName string `json:"member_name"`
		ChatCount  int    `json:"chat_count"`
	}, 0, len(familyChatCounts))

	for userID, count := range familyChatCounts {
		name := userID
		if nickname, ok := userNames[userID]; ok {
			name = nickname
		}
		familyInteractions = append(familyInteractions, struct {
			MemberName string `json:"member_name"`
			ChatCount  int    `json:"chat_count"`
		}{
			MemberName: name,
			ChatCount:  count,
		})
	}

	sort.Slice(familyInteractions, func(i, j int) bool {
		return familyInteractions[i].ChatCount > familyInteractions[j].ChatCount
	})

	return &SocialSummaryData{
		FamilyInteractions: familyInteractions,
		FriendAddedCount:   friendAddedCount,
		FriendChatCount:    friendChatCount,
	}
}

// stringToInt64Slice 将字符串切片转换为int64切片
func stringToInt64Slice(strs []string) []int64 {
	result := make([]int64, 0, len(strs))
	for _, s := range strs {
		var i int64
		if _, err := fmt.Sscanf(s, "%d", &i); err == nil && i > 0 {
			result = append(result, i)
		}
	}
	return result
}
