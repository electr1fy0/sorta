package llm

import (
	"context"
)

const DefaultModel = "groq/compound-mini"

type Request struct {
	Model        string
	SystemPrompt string
	UserPrompt   string
}

type Client interface {
	Run(ctx context.Context, req Request) (string, error)
}
