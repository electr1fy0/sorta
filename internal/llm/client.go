package llm

import (
	"context"
)

const DefaultModel = "openai/gpt-oss-20b"

type Request struct {
	Model        string
	SystemPrompt string
	UserPrompt   string
}

type Client interface {
	Run(ctx context.Context, req Request) (string, error)
}
