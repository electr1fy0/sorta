package internal

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var LogCnt = 0

// type Transaction struct {
// 	Operations []FileOperation
// 	ID         string
// 	Root       string
// }

func (e *Executor) RevertExecute(op FileOperation) error {
	srcDir := filepath.Dir(op.SourcePath)
	op.DestPath, op.SourcePath = op.SourcePath, op.DestPath

	if err := os.MkdirAll(srcDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	if err := os.Rename(op.SourcePath, op.DestPath); err != nil {
		return fmt.Errorf("failed to revert operation: %w", err)
	}
	return nil
}

func (e *Executor) Execute(op FileOperation) (bool, error) {
	if e.DryRun {
		return false, nil
	}

	switch op.Type {
	case OpMove:
		if op.DestPath == op.SourcePath {
			return false, nil
		}
		srcDir := filepath.Dir(op.SourcePath)
		srcDirName := filepath.Base(srcDir)
		if slices.Contains(e.Blacklist, srcDirName) {
			return false, nil
		}
		reader := bufio.NewReader(os.Stdin)
		if e.Interactive {
			fmt.Printf("[?] Move file \"%s\"? [y/n]: ", op.Filename)
			input, err := reader.ReadString('\n')
			if err != nil {
				return false, fmt.Errorf("error reading input: %w", err)
			}
			if strings.TrimSpace(input) != "y" {
				return false, nil
			}
		}

		destDir := filepath.Dir(op.DestPath)

		if err := os.MkdirAll(destDir, 0755); err != nil {
			return false, fmt.Errorf("failed to create directory: %w", err)
		}
		if err := os.Rename(op.SourcePath, op.DestPath); err != nil {
			return false, fmt.Errorf("failed to move file: %w", err)
		}

		if e.Interactive {
			fmt.Println("[?] Undo? [y/n]")
			undoInput, err := reader.ReadString('\n')
			if err != nil {
				e.Operations = append(e.Operations, op)
				return true, fmt.Errorf("error reading undo input: %w", err)
			}
			if strings.TrimSpace(undoInput) == "y" {
				if err := e.RevertExecute(op); err != nil {
					e.Operations = append(e.Operations, op)
					return true, fmt.Errorf("failed to undo move: %w", err)
				}
				return false, nil
			}
		}

		e.Operations = append(e.Operations, op)
		return true, nil

	case OpDelete:
		if err := os.Remove(op.SourcePath); err != nil {
			return false, fmt.Errorf("failed to delete file: %w", err)
		}
		return true, nil
	}

	return false, nil
}
