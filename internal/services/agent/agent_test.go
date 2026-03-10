package agent

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"aibuddy/pkg/config"
	"aibuddy/pkg/helpers"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	config.Setup("../../../config")

	// 初始化数据库连接
	query.SetDefault(model.Conn().GetDB())

	// 确保配置加载成功
	if config.Instance == nil || config.Instance.Agent == nil || config.Instance.Agent.Model == nil {
		fmt.Fprintf(os.Stderr, "配置加载失败: Agent 或 Model 为 nil\n")
		os.Exit(1)
	}
	if config.Instance.Agent.Model.ChatModel == nil {
		fmt.Fprintf(os.Stderr, "配置加载失败: ChatModel 为 nil\n")
		os.Exit(1)
	}

	helpers.PP(config.Instance.Agent.Model.ChatModel)

	os.Exit(m.Run())
}

// TestRoleChatAgentIntegration 集成测试，需要真实 LLM API
// 运行前请确保配置了正确的 API Key
func TestRoleChatAgentIntegration(t *testing.T) {
	service := NewRoleAgentService()

	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	endDate := time.Date(2026, 3, 31, 23, 59, 59, 0, time.Local)

	result, err := service.RoleChatAgent("30:ED:A0:E9:F3:13", startDate, endDate, "哲思少言型奶龙")
	require.NoError(t, err)
	require.NotNil(t, result)

	g, err := json.Marshal(result)
	assert.NoError(t, err)

	t.Logf("Result JSON: %s", string(g))
}

// TestInteractionSummary 测试角色互动小结
func TestInteractionSummary(t *testing.T) {
	report := NewGroupReport()

	// 测试时间范围
	startTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	endTime := time.Date(2026, 3, 31, 23, 59, 59, 0, time.Local)

	result := report.GetInteractionSummary("30:ED:A0:E9:F3:13", startTime, endTime)
	require.NotNil(t, result)

	// 验证返回结构
	assert.NotNil(t, result.TopRoles, "TopRoles should not be nil")

	// 打印结果
	g, err := json.MarshalIndent(result, "", "  ")
	require.NoError(t, err)
	t.Logf("InteractionSummary Result:\n%s", string(g))
}

// TestGetSocialSummary 测试社交互动总结
func TestGetSocialSummary(t *testing.T) {
	report := NewGroupReport()

	// 测试时间范围
	startTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	endTime := time.Date(2026, 3, 31, 23, 59, 59, 0, time.Local)

	result := report.GetSocialSummary("30:ED:A0:E9:F3:13", startTime, endTime)
	require.NotNil(t, result)

	// 验证返回结构
	assert.NotNil(t, result.FamilyInteractions, "FamilyInteractions should not be nil")

	// 打印结果
	g, err := json.MarshalIndent(result, "", "  ")
	require.NoError(t, err)
	t.Logf("GetSocialSummary Result:\n%s", string(g))
}
