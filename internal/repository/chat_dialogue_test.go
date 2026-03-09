package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"aibuddy/pkg/config"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/assert"

	"testing"
)

// cleanJSONResponse 清理模型响应中的 markdown 代码块标记
func cleanJSONResponse(content string) string {
	content = strings.TrimSpace(content)
	// 移除 ```json 或 ``` 开头
	content, _ = strings.CutPrefix(content, "```json")
	content, _ = strings.CutPrefix(content, "```")
	// 移除 ``` 结尾
	content, _ = strings.CutSuffix(content, "```")
	return strings.TrimSpace(content)
}

// extractJSON 从响应中提取有效的 JSON 对象
func extractJSON(content string) string {
	content = cleanJSONResponse(content)

	// 尝试找到 JSON 对象的起始位置
	startIdx := strings.Index(content, "{")
	if startIdx == -1 {
		return content
	}

	// 从起始位置开始，找到匹配的闭合括号
	braceCount := 0
	endIdx := -1
	for i := startIdx; i < len(content); i++ {
		if content[i] == '{' {
			braceCount++
		} else if content[i] == '}' {
			braceCount--
			if braceCount == 0 {
				endIdx = i + 1
				break
			}
		}
	}

	if endIdx > startIdx {
		return content[startIdx:endIdx]
	}

	return content
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// setupTest 初始化测试环境
func setupTest() {
	config.Setup("../../config")
	query.SetDefault(model.Conn().GetDB())
}

// ChatDialogueData 用于解析生成的对话数据
type ChatDialogueData struct {
	DeviceID     string `json:"device_id"`
	Question     string `json:"question"`
	QuestionTime string `json:"question_time"`
	Answer       string `json:"answer"`
	AnswerTime   string `json:"answer_time"`
	AgentName    string `json:"agent_name"`
}

// Test_GenerateTestData 生成单条测试数据
func Test_GenerateTestData(t *testing.T) {
	setupTest()
	ctx := context.Background()
	cfg := config.Instance
	configModel := cfg.Agent.Model.ChatModel
	modelName := configModel.ModelName
	apiKey := configModel.APIKey

	// 创建 ChatModel
	chatModel, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
		Model:  modelName,
		APIKey: apiKey,
	})
	assert.NoError(t, err, "创建 ChatModel 失败")

	// 当前时间
	now := time.Now()
	currentTime := now.Format("2006-01-02 15:04:05")

	// 构造 prompt，要求模型返回 JSON 格式的对话数据
	prompt := fmt.Sprintf(`你是一个测试数据生成器。请生成一条真实的用户与AI助手之间的对话数据。

当前时间是：%s

要求：
1. 问题应该是用户可能提出的真实问题
2. 回答应该专业、有帮助
3. 时间应该在当前时间附近（最近几天内）
4. 设备ID使用 dev_ 开头
5. agent_name（角色名称）必须从以下列表中随机选择一个：奶龙英语老师、奶龙斯坦、哲思少言型奶龙、阳光元气型奶龙、古灵精怪型奶龙

请严格按照以下 JSON 格式返回，不要包含任何其他内容：
{
  "device_id": "设备ID",
  "question": "问题内容",
  "question_time": "问题时间，格式：2006-01-02 15:04:05",
  "answer": "回答内容",
  "answer_time": "回答时间，格式：2006-01-02 15:04:05",
  "agent_name": "角色名称"
}

现在请生成一条聊天对话测试数据：`, currentTime)

	// 调用模型生成数据
	response, err := chatModel.Generate(ctx, []*schema.Message{
		schema.UserMessage(prompt),
	})
	assert.NoError(t, err, "生成数据失败")

	t.Logf("生成的原始响应: %s", response.Content)

	// 解析 JSON 数据
	jsonContent := extractJSON(response.Content)
	t.Logf("提取的 JSON: %s", jsonContent)

	var dialogueData ChatDialogueData
	err = json.Unmarshal([]byte(jsonContent), &dialogueData)
	assert.NoError(t, err, "解析 JSON 失败")

	// 解析时间
	questionTime, err := time.Parse("2006-01-02 15:04:05", dialogueData.QuestionTime)
	assert.NoError(t, err, "解析问题时间失败")

	answerTime, err := time.Parse("2006-01-02 15:04:05", dialogueData.AnswerTime)
	assert.NoError(t, err, "解析回答时间失败")

	// 创建对话记录
	dialogue := &model.ChatDialogue{
		DeviceID:     dialogueData.DeviceID,
		Question:     dialogueData.Question,
		QuestionTime: questionTime,
		Answer:       dialogueData.Answer,
		AnswerTime:   answerTime,
		AgentName:    dialogueData.AgentName,
	}

	// 保存到数据库
	err = query.ChatDialogue.Create(dialogue)
	assert.NoError(t, err, "保存数据失败")

	t.Logf("成功生成并保存对话数据:")
	t.Logf("  设备ID: %s", dialogue.DeviceID)
	t.Logf("  问题: %s", dialogue.Question)
	t.Logf("  问题时间: %s", dialogue.QuestionTime.Format("2006-01-02 15:04:05"))
	t.Logf("  回答: %s", dialogue.Answer)
	t.Logf("  回答时间: %s", dialogue.AnswerTime.Format("2006-01-02 15:04:05"))
	t.Logf("  角色名称: %s", dialogue.AgentName)
}

// Test_GenerateMultipleTestData 批量生成测试数据
func Test_GenerateMultipleTestData(t *testing.T) {
	setupTest()
	ctx := context.Background()
	cfg := config.Instance
	configModel := cfg.Agent.Model.ChatModel
	modelName := configModel.ModelName
	apiKey := configModel.APIKey

	// 创建 ChatModel
	chatModel, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
		Model:  modelName,
		APIKey: apiKey,
	})
	assert.NoError(t, err, "创建 ChatModel 失败")

	// 当前时间
	now := time.Now()
	currentTime := now.Format("2006-01-02 15:04:05")

	count := 20
	prompt := fmt.Sprintf(`你是一个测试数据生成器。请生成 %d 条真实的用户与AI助手之间的对话数据。

当前时间是：%s

要求：
1. 问题应该多样化，涵盖不同主题
2. 回答应该专业、有帮助
3. 时间应该在当前时间附近（最近几天内）
4. 每条数据的设备ID不同，使用 30:ED:A0:E9:F3:06, 30:ED:A0:E9:F3:05 这样的格式
5. agent_name（角色名称）必须从以下列表中随机选择：奶龙英语老师、奶龙斯坦、哲思少言型奶龙、阳光元气型奶龙、古灵精怪型奶龙

请严格按照以下 JSON 格式返回，不要包含任何其他内容：
{
  "dialogues": [
    {
      "device_id": "设备ID",
      "question": "问题内容",
      "question_time": "问题时间",
      "answer": "回答内容",
      "answer_time": "回答时间",
      "agent_name": "角色名称"
    }
  ]
}

现在请生成 %d 条聊天对话测试数据：`, count, currentTime, count)

	// 调用模型生成数据
	response, err := chatModel.Generate(ctx, []*schema.Message{
		schema.UserMessage(prompt),
	})
	assert.NoError(t, err, "生成数据失败")

	t.Logf("生成的原始响应: %s", response.Content)

	// 解析批量数据
	jsonContent := extractJSON(response.Content)
	t.Logf("提取的 JSON 长度: %d 字符", len(jsonContent))

	var batchData struct {
		Dialogues []ChatDialogueData `json:"dialogues"`
	}
	err = json.Unmarshal([]byte(jsonContent), &batchData)
	if err != nil {
		// 打印更详细的错误信息
		t.Logf("JSON 解析错误: %v", err)
		t.Logf("问题 JSON 内容 (前500字符): %s", truncateString(jsonContent, 500))
	}
	assert.NoError(t, err, "解析 JSON 失败")

	// 批量保存
	var dialogues []*model.ChatDialogue
	for _, d := range batchData.Dialogues {
		questionTime, err := time.Parse("2006-01-02 15:04:05", d.QuestionTime)
		if err != nil {
			t.Logf("解析问题时间失败: %v, 跳过此条", err)
			continue
		}

		answerTime, err := time.Parse("2006-01-02 15:04:05", d.AnswerTime)
		if err != nil {
			t.Logf("解析回答时间失败: %v, 跳过此条", err)
			continue
		}

		dialogues = append(dialogues, &model.ChatDialogue{
			DeviceID:     d.DeviceID,
			Question:     d.Question,
			QuestionTime: questionTime,
			Answer:       d.Answer,
			AnswerTime:   answerTime,
			AgentName:    d.AgentName,
		})
	}

	if len(dialogues) > 0 {
		err = query.ChatDialogue.CreateInBatches(dialogues, 100)
		assert.NoError(t, err, "批量保存数据失败")
		t.Logf("成功生成并保存 %d 条对话数据", len(dialogues))
	}
}
