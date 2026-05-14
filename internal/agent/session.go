package agent

import (
	"context"

	"github.com/electr1fy0/sorta/internal/llm"
)

type SessionOptions struct {
	MaxSteps int
}

type Session struct {
	client   llm.Client
	maxSteps int
}

func NewSession(client llm.Client, opts SessionOptions) *Session {
	maxSteps := opts.MaxSteps
	if maxSteps <= 0 {
		maxSteps = 3
	}
	return &Session{client: client, maxSteps: maxSteps}
}

func (s *Session) Plan(ctx context.Context, dir, instruction, model string, hints []string) (ExecutionPlan, error) {
	snapshot, err := SummarizeDirectory(ctx, dir)
	if err != nil {
		return ExecutionPlan{}, err
	}

	userPrompt, err := BuildUserPrompt(PromptInput{
		Directory:   dir,
		Instruction: instruction,
		Observed:    snapshot,
		UserHints:   append([]string(nil), hints...),
	})
	if err != nil {
		return ExecutionPlan{}, err
	}

	raw, err := s.client.Run(ctx, llm.Request{
		Model:        model,
		SystemPrompt: GoalSystemPrompt(),
		UserPrompt:   userPrompt,
	})
	if err != nil {
		return ExecutionPlan{}, err
	}

	return ParseExecutionPlan(raw)
}
