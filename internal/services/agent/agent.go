// Package agent provides the agent service.
package agent

import (
	"aibuddy/pkg/config"

	"github.com/aws/smithy-go/ptr"
	"trpc.group/trpc-go/trpc-agent-go/agent/llmagent"
	"trpc.group/trpc-go/trpc-agent-go/model"
	"trpc.group/trpc-go/trpc-agent-go/model/openai"
)

// NewAgentModel 创建一个新的LLMAgent模型
func NewAgentModel(output any, description string) *llmagent.LLMAgent {
	configModel := config.Instance.Agent.Model.ChatModel
	modelName := configModel.ModelName
	apiKey := configModel.APIKey
	baseURL := configModel.APIURL

	modelInstance := openai.New(modelName,
		openai.WithAPIKey(apiKey),
		openai.WithBaseURL(baseURL),
		openai.WithExtraFields(map[string]any{
			"enable_search":    true,
			"web_extractor":    true,
			"code_interpreter": true,
		}),
	)
	agent := llmagent.New(
		"aibuddy_agent",
		llmagent.WithModel(modelInstance),
		llmagent.WithGenerationConfig(
			model.GenerationConfig{
				Stream:          false,
				ThinkingEnabled: ptr.Bool(true),
			},
		),
		llmagent.WithDescription(description),
		llmagent.WithStructuredOutputJSON(output, true, "请严格按照这个JSON Schema输出,不要输出任何其他内容"),
	)
	return agent
}
