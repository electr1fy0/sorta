package agent

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/electr1fy0/sorta/internal/config"
	"github.com/electr1fy0/sorta/internal/core"
	"github.com/electr1fy0/sorta/internal/dupl"
	"github.com/electr1fy0/sorta/internal/ignore"
	"github.com/electr1fy0/sorta/internal/llm"
	"github.com/electr1fy0/sorta/internal/ops"
	"github.com/electr1fy0/sorta/internal/rename"
	"github.com/electr1fy0/sorta/internal/sorter"
)

func ExecutePlan(ctx context.Context, dir string, plan ExecutionPlan) error {
	for i := 0; i < len(plan.Actions); i++ {
		action := plan.Actions[i]
		switch action.Kind {
		case ActionDedupe:
			if err := runSorter(ctx, dir, dupl.NewDuplicateFinder(), nil); err != nil {
				return err
			}
		case ActionRename:
			sorter, err := newRenameSelectionSorter(action.Rename.Files)
			if err != nil {
				return err
			}
			if err := runSorter(ctx, dir, sorter, nil); err != nil {
				return err
			}
		case ActionSortRule:
			rules := make([]config.RuleSpec, 0, 4)
			for ; i < len(plan.Actions) && plan.Actions[i].Kind == ActionSortRule; i++ {
				rule := plan.Actions[i].SortRule
				rules = append(rules, config.RuleSpec{
					Folder:   rule.Folder,
					Keywords: append([]string(nil), rule.Keywords...),
				})
			}
			i--

			ruleSorter, err := sorter.NewRuleSorter(rules)
			if err != nil {
				return err
			}
			if err := runSorter(ctx, dir, ruleSorter, ruleSorter.GetBlacklist()); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported planned action: %s", action.Kind)
		}
	}

	return nil
}

type renameSelectionSorter struct {
	renamer *rename.Renamer
	allowed map[string]struct{}
}

func newRenameSelectionSorter(files []string) (*renameSelectionSorter, error) {
	client, err := llm.NewClient(llm.DefaultModel)
	if err != nil {
		return nil, err
	}
	allowed := make(map[string]struct{}, len(files))
	for _, file := range files {
		allowed[filepath.Clean(file)] = struct{}{}
	}
	return &renameSelectionSorter{
		renamer: rename.NewRenamer(client, llm.DefaultModel, nil),
		allowed: allowed,
	}, nil
}

func (s *renameSelectionSorter) Decide(ctx context.Context, files []core.FileEntry) ([]core.FileOperation, error) {
	if len(s.allowed) == 0 {
		return s.renamer.Decide(ctx, files)
	}

	selected := make([]core.FileEntry, 0, len(files))
	ops := make([]core.FileOperation, 0, len(files))
	for _, file := range files {
		rel, err := filepath.Rel(file.RootDir, file.SourcePath)
		if err != nil {
			return nil, err
		}
		if _, ok := s.allowed[filepath.Clean(rel)]; ok {
			selected = append(selected, file)
			continue
		}
		ops = append(ops, core.FileOperation{OpType: core.OpSkip, File: file})
	}

	renames, err := s.renamer.Decide(ctx, selected)
	if err != nil {
		return nil, err
	}
	ops = append(ops, renames...)
	return ops, nil
}

func runSorter(ctx context.Context, dir string, sorter core.Sorter, ignorePatterns []string) error {
	ignoreMatcher, err := ignore.LoadIgnoreMatcher(dir, ignorePatterns)
	if err != nil {
		return fmt.Errorf("failed to load ignore patterns: %w", err)
	}

	plannedOps, err := ops.PlanOperationsWithIgnoreCtx(ctx, dir, sorter, ignoreMatcher)
	if err != nil {
		return fmt.Errorf("failed to plan operations: %w", err)
	}

	cleanedOps := make([]core.FileOperation, 0, len(plannedOps))
	for _, op := range plannedOps {
		if op.DestPath == op.File.SourcePath {
			continue
		}
		cleanedOps = append(cleanedOps, op)
	}

	if len(cleanedOps) == 0 {
		fmt.Println("No operations needed.")
		return nil
	}

	executor := &ops.Executor{Operations: make([]core.FileOperation, 0, len(cleanedOps))}
	reporter := &ops.Reporter{}
	res, err := ops.ApplyOperationsCtx(ctx, dir, cleanedOps, executor, reporter)
	if err != nil {
		return fmt.Errorf("failed to apply operations: %w", err)
	}
	res.PrintSummary()
	return nil
}
