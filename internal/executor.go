package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

func (e *Executor) Execute(op FileOperation) (bool, error) {
	if e.DryRun {
		return false, nil
	}

	switch op.OpType {
	case OpMove, OpDedupe, OpRename:
		if op.DestPath == op.File.SourcePath {
			return false, nil
		}
		srcDir := filepath.Dir(op.File.SourcePath)
		srcDirName := filepath.Base(srcDir)
		if slices.Contains(e.Blacklist, srcDirName) {
			return false, nil
		}
		destDir := filepath.Dir(op.DestPath)

		if err := os.MkdirAll(destDir, 0755); err != nil {
			return false, fmt.Errorf("failed to create directory: %w", err)
		}
		if err := os.Rename(op.File.SourcePath, op.DestPath); err != nil {
			return false, fmt.Errorf("failed to move file: %w", err)
		}

		e.Operations = append(e.Operations, op)
		return true, nil

	case OpDelete:
		if err := os.Remove(op.File.SourcePath); err != nil {
			return false, fmt.Errorf("failed to delete file: %w", err)
		}
		return true, nil
	}

	return false, nil
}
