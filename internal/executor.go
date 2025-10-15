package internal

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var Operations = make([]FileOperation, 10)

func (e *Executor) RevertExecute(op FileOperation) error {
	srcDir := filepath.Dir(op.SourcePath)

	if err := os.MkdirAll(srcDir, 0755); err != nil {
		return err
	}
	if err := os.Rename(op.DestPath, op.SourcePath); err != nil {
		return err
	}
	return nil
}

func (e *Executor) Execute(op FileOperation) (bool, error) {
	if e.DryRun {
		return false, nil
	}

	switch op.Type {
	case OpMove:
		destDir := filepath.Dir(op.DestPath)

		if op.DestPath == op.SourcePath {
			return false, nil
		}

		reader := bufio.NewReader(os.Stdin)
		if e.Interactive && op.Type == OpMove {
			fmt.Printf("[?] Move file \"%s\"? [y/n]: ", op.Filename)
			input, err := reader.ReadString('\n')
			if err != nil {
				return false, err
			}
			if strings.TrimSpace(input) != "y" {
				return false, nil
			}
		}

		if err := os.MkdirAll(destDir, 0755); err != nil {
			return false, err
		}
		if err := os.Rename(op.SourcePath, op.DestPath); err != nil {
			return false, fmt.Errorf("error executing: %w", err)
		}
		if e.Interactive {
			fmt.Println("[?] Undo? [y/n]")
			undoInput, err := reader.ReadString('\n')
			if err != nil {
				return false, err
			}
			if strings.TrimSpace(undoInput) == "y" {
				e.RevertExecute(op)
			}

		}
		Operations = append(Operations, op)

		return true, nil

	case OpDelete:
		if err := os.Remove(op.SourcePath); err != nil {
			return false, err
		}
	}

	return false, nil
}
