package agent

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"aibuddy/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestEmotionWarningService_GenerateWarning(t *testing.T) {
	service := NewEmotionWarningService()

	dialogues := []*model.ChatDialogue{
		{
			Question:     "他们都不理我，我真的不想去学校了",
			QuestionTime: time.Now(),
			Answer:       "听起来你最近遇到了一些困难，愿意和我说说发生了什么吗？",
			AnswerTime:   time.Now().Add(time.Second),
		},
		{
			Question:     "我最近一直睡不好，越想越难受",
			QuestionTime: time.Now().Add(time.Minute),
			Answer:       "睡眠问题确实会影响心情，这种情况持续多久了？",
			AnswerTime:   time.Now().Add(time.Minute + time.Second),
		},
	}

	result, err := service.GenerateWarning(dialogues)
	if err != nil {
		t.Fatalf("GenerateWarning failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// 打印模型输出结果
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatalf("Marshal result failed: %v", err)
	}
	t.Logf("Model Output:\n%s", string(jsonData))

	// 验证关键字段
	if !result.TriggerWarning {
		t.Error("Expected TriggerWarning true")
	}
	if result.WarningLevel == "" {
		t.Error("Expected non-empty WarningLevel")
	}
	if len(result.WarningTypes) == 0 {
		t.Error("Expected at least one WarningType")
	}
	if len(result.Evidence) == 0 {
		t.Error("Expected at least one Evidence")
	}
	if len(result.ParentSuggestions) < 3 {
		t.Error("Expected at least 3 ParentSuggestions")
	}

	emotion, err := result.ToEmotion("30:ED:A0:E9:F3:03", 123)
	assert.NoError(t, err)
	err = service.CreateEmotion(context.Background(), emotion)
	assert.NoError(t, err)
}
