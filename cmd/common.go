package cmd

import (
	"fmt"
	"os"
	"path/filepath"

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
	home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine home directory: %w", err)
		}
		path = filepath.Join(home, path)
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

	executor := &internal.Executor{
		DryRun:     dryRun,
		Blacklist:  blacklist,
		Operations: make([]internal.FileOperation, 0),
	}
	reporter := &internal.Reporter{DryRun: dryRun}

	res, err := internal.FilterFiles(dir, sorter, executor, reporter)
	if err != nil {
		return fmt.Errorf("failed to filter files: %w", err)
	}

	res.PrintSummary()

	return nil
}