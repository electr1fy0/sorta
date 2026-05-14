package llm

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type PlannerClient struct {
	client       *openai.Client
	defaultModel string
}

func NewClient(model string) (*PlannerClient, error) {
	key := strings.TrimSpace(os.Getenv("GROQ_API_KEY"))
	if key == "" {
		return nil, fmt.Errorf("missing API credentials: set GROQ_API_KEY")
	}

	client := openai.NewClient(
		option.WithAPIKey(key),
		option.WithBaseURL("https://api.groq.com/openai/v1/"),
	)

	return &PlannerClient{
		client:       &client,
		defaultModel: model,
	}, nil
}

func (c *PlannerClient) Run(ctx context.Context, req Request) (string, error) {
	model := req.Model
	if strings.TrimSpace(req.Model) == "" {
		model = c.defaultModel
	}

	params := openai.ChatCompletionNewParams{
		Model: openai.ChatModel(model),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(req.SystemPrompt),
			openai.UserMessage(req.UserPrompt),
		},
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONObject: &openai.ResponseFormatJSONObjectParam{
				Type: "json_object",
			},
		},
	}

	resp, err := c.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("planner returned no choices")
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}
