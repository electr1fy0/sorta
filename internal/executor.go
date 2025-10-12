package internal

import (
	"fmt"
	"os"
	"path/filepath"
)

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
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return false, err
		}
		err := os.Rename(op.SourcePath, op.DestPath)

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
