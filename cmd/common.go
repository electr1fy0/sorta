package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/electr1fy0/sorta/internal"
)

const (
	ansiReset = "[0m"
	ansiCyan  = "[36m"
)

func resolvePath(path string) (string, error) {
	var err error
	path, err = internal.ExpandPath(path)
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

func runSort(dir string, sorter internal.Sorter, blacklist []string) error {
	fmt.Printf("%sDir:%s %s\n", ansiCyan, ansiReset, dir)
	fmt.Println("Analyzing files...")

	ops, err := internal.PlanOperations(dir, sorter)
	if err != nil {
		return fmt.Errorf("failed to plan operations: %w", err)
	}

	if len(ops) == 0 {
		fmt.Println("No operations needed.")
		return nil
	}

	moves := 0
	deletes := 0
	skips := 0
	for _, op := range ops {
		switch op.OpType {
		case internal.OpMove:
			moves++
		case internal.OpDelete:
			deletes++
		case internal.OpSkip:
			skips++
		}
	}

	fmt.Printf("Found %d operations:\n", len(ops))
	if moves > 0 {
		fmt.Printf("- %d files to move\n", moves)
	}
	if deletes > 0 {
		fmt.Printf("- %d files to delete\n", deletes)
	}
	if skips > 0 {
		fmt.Printf("- %d files skipped (no match)\n", skips)
	}

	if dryRun {
		fmt.Println("\nDry run complete. No changes made.")
		return nil
	}

	fmt.Print("\nDo you want to proceed? [y/N]: ")
	reader := bufio.NewReader(os.Stdin)
	ans, _ := reader.ReadString('\n')
	if strings.ToLower(strings.TrimSpace(ans)) != "y" {
		fmt.Println("Operation cancelled.")
		return nil
	}

	executor := &internal.Executor{
		DryRun:     false,
		Blacklist:  blacklist,
		Operations: make([]internal.FileOperation, 0),
	}
	reporter := &internal.Reporter{DryRun: false}

	res, err := internal.ApplyOperations(dir, ops, executor, reporter)
	if err != nil {
		return fmt.Errorf("failed to apply operations: %w", err)
	}

	res.PrintSummary()

	return nil
}
