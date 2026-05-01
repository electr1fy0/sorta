package agent

import (
	"context"

	"github.com/electr1fy0/sorta/internal/llm"
)

type SessionOptions struct {
	MaxSteps int
}

type PlanInput struct {
	Dir         string
	Instruction string
	Model       string
	Names       NamesInput
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

type PlanResult struct {
	Plan             ExecutionPlan
	RequestFilenames []string
}

func (s *Session) Plan(ctx context.Context, input PlanInput) (PlanResult, error) {
	snapshot, err := SummarizeDirectory(ctx, input.Dir)
	if err != nil {
		return PlanResult{}, err
	}

	userPrompt, err := BuildUserPrompt(PromptInput{
		Directory:   input.Dir,
		Instruction: input.Instruction,
		Observed:    snapshot,
		UserHints:   append([]string(nil), input.Names.Hints...),
	})
	if err != nil {
		return PlanResult{}, err
	}

	raw, err := s.client.Run(ctx, llm.Request{
		Model:        input.Model,
		SystemPrompt: GoalSystemPrompt(),
		UserPrompt:   userPrompt,
	})
	if err != nil {
		return PlanResult{}, err
	}

	plan, err := ParseExecutionPlan(raw)
	if err != nil {
		return PlanResult{}, err
	}

	return PlanResult{
		Plan:             plan,
		RequestFilenames: append([]string(nil), plan.RequestFilenames...),
	}, nil
}
