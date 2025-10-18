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

func validateDir(path string) (string, error) {
	if filepath.IsAbs(path) {
		path = filepath.Clean(path)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine home directory: %w", err)
		}
		path = filepath.Join(home, path)
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

func runSort(dir string, sorter internal.Sorter) error {
	fmt.Printf("%sDir:%s %s\n", ansiCyan, ansiReset, dir)

	executor := &internal.Executor{DryRun: dryRun, Interactive: interactive}
	reporter := &internal.Reporter{DryRun: dryRun}

	res, err := internal.FilterFiles(dir, sorter, executor, reporter)
	if err != nil {
		return fmt.Errorf("failed to filter files: %w", err)
	}

	if err := Undo(); err != nil {
		return err
	}
	res.PrintSummary()

	return nil
}

var undoExecutor = internal.Executor{Interactive: interactive, DryRun: dryRun}

func Undo() error {
	if undoExecutor.Interactive || dryRun {
		return nil
	}

	if len(internal.Operations) == 0 {
		fmt.Println("No operations to undo.")
		return nil
	}

	fmt.Println("[?] Undo? [y/n]")
	input := bufio.NewReader(os.Stdin)
	confirm, err := input.ReadString('\n')
	if err != nil {
		return fmt.Errorf("error taking undo input: %w", err)
	}

	if strings.TrimSpace(confirm) == "y" {
		for _, op := range internal.Operations {
			if err := undoExecutor.RevertExecute(op); err != nil {
				return fmt.Errorf("error reverting operation: %w", err)
			}
		}
		fmt.Println("Changes reverted.")
	}
	return nil
}
