package agent

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/electr1fy0/sorta/internal/config"
	"github.com/electr1fy0/sorta/internal/core"
	"github.com/electr1fy0/sorta/internal/dupl"
	"github.com/electr1fy0/sorta/internal/rename"
	"github.com/electr1fy0/sorta/internal/sorter"
)

type ActionKind string

const (
	ActionSortRule ActionKind = "sort_rule"
	ActionDedupe   ActionKind = "dedupe"
	ActionRename   ActionKind = "rename"
	ActionMkdir    ActionKind = "mkdir"
)

type SortRuleAction struct {
	Folder   string   `json:"folder"`
	Keywords []string `json:"keywords"`
}

type RenameAction struct {
	Files []string `json:"files,omitempty"`
	Hints []string `json:"hints,omitempty"`
}

type MkdirAction struct {
	Path string `json:"path"`
}

type PlannedAction struct {
	Kind     ActionKind      `json:"kind"`
	SortRule *SortRuleAction `json:"sort_rule,omitempty"`
	Rename   *RenameAction   `json:"rename,omitempty"`
	Mkdir    *MkdirAction    `json:"mkdir,omitempty"`
}

type ExecutionPlan struct {
	Summary          string          `json:"summary"`
	RequestFilenames []string        `json:"request_filenames,omitempty"`
	Actions          []PlannedAction `json:"actions"`
}

func ParseExecutionPlan(raw string) (ExecutionPlan, error) {
	var plan ExecutionPlan
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &plan); err != nil {
		return ExecutionPlan{}, fmt.Errorf("failed to parse plan JSON: %w", err)
	}
	if err := plan.Validate(); err != nil {
		return ExecutionPlan{}, err
	}
	return plan, nil
}

func (p *ExecutionPlan) Validate() error {
	p.Summary = strings.TrimSpace(p.Summary)
	questions := make([]string, 0, len(p.RequestFilenames))
	for _, question := range p.RequestFilenames {
		question = strings.TrimSpace(question)
		if question != "" {
			questions = append(questions, question)
		}
	}
	p.RequestFilenames = questions
	if p.Actions == nil {
		p.Actions = []PlannedAction{}
	}
	if len(p.RequestFilenames) > 0 && len(p.Actions) > 0 {
		return fmt.Errorf("plan cannot contain both request_filenames and actions")
	}
	for i := range p.Actions {
		if err := p.Actions[i].normalize(); err != nil {
			return fmt.Errorf("action %d: %w", i+1, err)
		}
	}
	return nil
}

func (a *PlannedAction) ToSorter() (core.Sorter, error) {
	switch a.Kind {
	case ActionDedupe:
		return dupl.NewDuplicateFinder(), nil
	case ActionRename:
		return rename.NewSelectionSorter(a.Rename.Files, a.Rename.Hints)
	case ActionSortRule:
		rules := []config.RuleSpec{{
			Folder:   a.SortRule.Folder,
			Keywords: append([]string(nil), a.SortRule.Keywords...),
		}}
		return sorter.NewRuleSorter(rules)
	default:
		return nil, fmt.Errorf("action kind %q cannot be converted to a sorter", a.Kind)
	}
}

func (a *PlannedAction) normalize() error {
	switch a.Kind {
	case ActionDedupe:
		return nil
	case ActionSortRule:
		if a.SortRule == nil {
			return fmt.Errorf("sort_rule payload is required")
		}
		a.SortRule.Folder = strings.TrimSpace(a.SortRule.Folder)
		if a.SortRule.Folder == "" {
			return fmt.Errorf("sort_rule.folder cannot be empty")
		}
		keywords := make([]string, 0, len(a.SortRule.Keywords))
		for _, keyword := range a.SortRule.Keywords {
			keyword = strings.TrimSpace(keyword)
			if keyword != "" {
				keywords = append(keywords, keyword)
			}
		}
		if len(keywords) == 0 {
			return fmt.Errorf("sort_rule.keywords cannot be empty")
		}
		a.SortRule.Keywords = keywords
		return nil
	case ActionRename:
		if a.Rename == nil {
			return fmt.Errorf("rename payload is required")
		}
		files := make([]string, 0, len(a.Rename.Files))
		for _, file := range a.Rename.Files {
			cleaned, err := normalizeRelativePath(file)
			if err != nil {
				return fmt.Errorf("invalid rename target %q: %w", file, err)
			}
			files = append(files, cleaned)
		}
		a.Rename.Files = files
		return nil
	case ActionMkdir:
		if a.Mkdir == nil {
			return fmt.Errorf("mkdir payload is required")
		}
		cleaned, err := normalizeRelativePath(a.Mkdir.Path)
		if err != nil {
			return fmt.Errorf("invalid mkdir path %q: %w", a.Mkdir.Path, err)
		}
		a.Mkdir.Path = cleaned
		return nil
	default:
		return fmt.Errorf("unsupported action kind %q", a.Kind)
	}
}

func (p ExecutionPlan) String() string {
	var b strings.Builder
	if strings.TrimSpace(p.Summary) != "" {
		fmt.Fprintf(&b, "Goal summary: %s\n", p.Summary)
	}
	if len(p.Actions) == 0 {
		b.WriteString("No actions planned.\n")
		return b.String()
	}

	b.WriteString("Planned actions:\n")
	for i, action := range p.Actions {
		switch action.Kind {
		case ActionSortRule:
			fmt.Fprintf(&b, "%d. sort into %q using [%s]\n", i+1, action.SortRule.Folder, strings.Join(action.SortRule.Keywords, ", "))
		case ActionDedupe:
			fmt.Fprintf(&b, "%d. remove duplicates\n", i+1)
		case ActionRename:
			if len(action.Rename.Files) == 0 {
				fmt.Fprintf(&b, "%d. rename files across the directory\n", i+1)
			} else {
				fmt.Fprintf(&b, "%d. rename selected files: %s\n", i+1, strings.Join(action.Rename.Files, ", "))
			}
		case ActionMkdir:
			fmt.Fprintf(&b, "%d. create directory %q\n", i+1, action.Mkdir.Path)
		default:
			fmt.Fprintf(&b, "%d. %s\n", i+1, action.Kind)
		}
	}

	return b.String()
}

func normalizeRelativePath(path string) (string, error) {
	cleaned := filepath.Clean(strings.TrimSpace(path))
	if cleaned == "." || cleaned == "" {
		return "", fmt.Errorf("path cannot be empty")
	}
	if filepath.IsAbs(cleaned) {
		return "", fmt.Errorf("path must be relative")
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path cannot escape the target directory")
	}
	return cleaned, nil
}
