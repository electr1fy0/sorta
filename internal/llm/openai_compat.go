package llm

import (
	"context"
	"fmt"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type PlannerClient struct {
	client       *openai.Client
	defaultModel string
}

func NewClient(model string) (*PlannerClient, error) {
	cfg, err := loadAPIConfig()
	if err != nil {
		return nil, err
	}

	cfg.BaseURL = "https://api.groq.com/openai/v1"

	return &PlannerClient{
		client:       openai.NewClientWithConfig(cfg),
		defaultModel: model,
	}, nil
}

func (c *PlannerClient) Run(ctx context.Context, req Request) (string, error) {
	model := req.Model
	if strings.TrimSpace(req.Model) == "" {
		model = c.defaultModel
	}

	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: req.SystemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: req.UserPrompt},
		},
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("planner returned no choices")
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

func loadAPIConfig() (openai.ClientConfig, error) {
	if key := strings.TrimSpace(os.Getenv("GROQ_API_KEY")); key != "" {
		return openai.DefaultConfig(key), nil
	}
	return openai.ClientConfig{}, fmt.Errorf("missing API credentials: set GROQ_API_KEY")
}
