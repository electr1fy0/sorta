package internal

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var Operations = make([]FileOperation, 10)

func (e *Executor) Execute(op FileOperation) (bool, error) {
	if e.DryRun {
		return false, nil
	}

	switch op.Type {
	case OpMove:
		destDir := filepath.Dir(op.DestPath)
		srcDir := filepath.Dir(op.SourcePath)

		if op.DestPath == op.SourcePath {
			return false, nil
		}

		if e.Undo {
			if err := os.MkdirAll(srcDir, 0755); err != nil {
				return false, err
			}

			err := os.Rename(op.DestPath, op.SourcePath)
			return false, err
		}

		if e.Interactive && op.Type == OpMove {
			fmt.Printf("[?] Move file \"%s\"? [y/n]: ", op.Filename)
			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				return false, err
			}

			choice := strings.TrimSpace(input)
			if choice != "y" {
				return false, nil
			}
		}

		var err error

		if err := os.MkdirAll(destDir, 0755); err != nil {
			return false, err
		}
		err = os.Rename(op.SourcePath, op.DestPath)
		Operations = append(Operations, op)
		if err != nil {
			return false, fmt.Errorf("error executing: %w", err)
		}
		return true, nil
	case OpDelete:
		if err := os.Remove(op.SourcePath); err != nil {
			return false, err
		}
	}

	return false, nil
}
