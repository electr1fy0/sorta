package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/electr1fy0/sorta/internal/core"
	"github.com/electr1fy0/sorta/internal/ignore"
	"github.com/electr1fy0/sorta/internal/ops"
	"github.com/electr1fy0/sorta/internal/tui"
	"github.com/mattn/go-isatty"
)

const (
	ansiReset = "[0m"
	ansiCyan  = "[36m"
)

func resolvePath(path string) (string, error) {
	var err error
	path, err = core.ExpandPath(path)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(path) {
		path, err = filepath.Abs(path)
		if err != nil {
			return "", fmt.Errorf("cannot determine absolute path: %w", err)
		}
	}
	return filepath.Clean(path), nil
}

func validateDir(path string) (string, error) {
	path, err := resolvePath(path)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("directory does not exist: %s", path)
		}
		return "", fmt.Errorf("cannot access path: %w", err)
	}

	if !info.IsDir() {
		return "", fmt.Errorf("not a directory: %s", path)
	}

	return path, nil
}

func runSort(dir string, sorter core.Sorter, ignorePatterns []string) error {
	fmt.Printf("%sDir:%s %s\n", ansiCyan, ansiReset, dir)
	fmt.Println("Analyzing files...")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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

	moves, deletes, skips, renames, dedupes := 0, 0, 0, 0, 0
	for _, op := range cleanedOps {
		switch op.OpType {
		case core.OpMove:
			moves++
		case core.OpDelete:
			deletes++
		case core.OpSkip:
			skips++
		case core.OpRename:
			renames++
		case core.OpDedupe:
			dedupes++
		}
	}

	fmt.Printf("Found %d operations:\n", len(cleanedOps))
	if moves > 0 {
		fmt.Printf("- %d files to move\n", moves)
	}
	if deletes > 0 {
		fmt.Printf("- %d files to delete\n", deletes)
	}
	if renames > 0 {
		fmt.Printf("- %d files to rename\n", renames)
	}
	if dedupes > 0 {
		fmt.Printf("- %d files to deduplicate\n", dedupes)
	}
	if skips > 0 {
		fmt.Printf("- %d files skipped (no match)\n", skips)
	}

	if dryRun {
		fmt.Println("\nDry run complete. No changes made.")
		return nil
	}

	if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		var tuiOps []core.FileOperation
		for _, op := range cleanedOps {
			if op.OpType != core.OpSkip {
				tuiOps = append(tuiOps, op)
			}
		}

		if len(tuiOps) == 0 {
			fmt.Println("No changes to make.")
			return nil
		}

		selectedOps, err := tui.SelectOperations(dir, tuiOps)
		if err != nil {
			fmt.Println("Operation cancelled.")
			return nil
		}
		cleanedOps = selectedOps
		if len(cleanedOps) == 0 {
			fmt.Println("No operations selected.")
			return nil
		}
	} else {
		fmt.Print("\nDo you want to proceed? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		ans, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(ans)) != "y" {
			fmt.Println("Operation cancelled.")
			return nil
		}
	}

	executor := &ops.Executor{
		Operations: make([]core.FileOperation, 0),
	}
	reporter := &ops.Reporter{}

	res, err := ops.ApplyOperationsCtx(ctx, dir, cleanedOps, executor, reporter)
	if err != nil {
		return fmt.Errorf("failed to apply operations: %w", err)
	}

	res.PrintSummary()

	return nil
}
