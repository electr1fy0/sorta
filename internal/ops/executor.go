package ops

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/electr1fy0/sorta/internal/core"
)

type Executor struct {
	Operations []core.FileOperation
}

func (e *Executor) Execute(op core.FileOperation) (bool, error) {
	switch op.OpType {
	case core.OpMove, core.OpDedupe, core.OpRename:
		if op.DestPath == op.File.SourcePath {
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

	case core.OpDelete:
		if err := os.Remove(op.File.SourcePath); err != nil {
			return false, fmt.Errorf("failed to delete file: %w", err)
		}
		return true, nil
	}

	return false, nil
}
