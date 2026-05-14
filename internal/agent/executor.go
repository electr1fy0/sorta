package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/electr1fy0/sorta/internal/config"
	"github.com/electr1fy0/sorta/internal/ops"
	"github.com/electr1fy0/sorta/internal/sorter"
)

func ExecutePlan(ctx context.Context, dir string, plan ExecutionPlan) error {
	for i := 0; i < len(plan.Actions); i++ {
		action := plan.Actions[i]

		if action.Kind == ActionMkdir {
			if err := os.MkdirAll(filepath.Join(dir, action.Mkdir.Path), 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}

		// batch consecutive sort_rule actions to avoid redundant directory walks
		if action.Kind == ActionSortRule {
			rules := make([]config.RuleSpec, 0, 4)
			for ; i < len(plan.Actions) && plan.Actions[i].Kind == ActionSortRule; i++ {
				rules = append(rules, config.RuleSpec{
					Folder:   plan.Actions[i].SortRule.Folder,
					Keywords: append([]string(nil), plan.Actions[i].SortRule.Keywords...),
				})
			}
			i--

			s, err := sorter.NewRuleSorter(rules)
			if err != nil {
				return err
			}
			if err := ops.RunSorter(ctx, dir, s, s.GetBlacklist()); err != nil {
				return err
			}
			continue
		}

		s, err := action.ToSorter()
		if err != nil {
			return err
		}
		if err := ops.RunSorter(ctx, dir, s, nil); err != nil {
			return err
		}
	}

	return nil
}
