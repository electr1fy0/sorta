package ops

import (
	"fmt"
	"path/filepath"

	"github.com/electr1fy0/sorta/internal/core"
)

type Executor struct {
	FS         core.FileSystem
	Operations []core.FileOperation
}

func (e *Executor) Execute(op core.FileOperation) (bool, error) {
	return e.ExecuteWithSrc(op, op.File.SourcePath)
}

func (e *Executor) ExecuteWithSrc(op core.FileOperation, src string) (bool, error) {
	if e.FS == nil {
		e.FS = core.OSFileSystem{}
	}

	switch op.OpType {
	case core.OpMove, core.OpDedupe, core.OpRename:
		if op.DestPath == src {
			return false, nil
		}
		destDir := filepath.Dir(op.DestPath)

		if err := e.FS.MkdirAll(destDir, 0755); err != nil {
			return false, fmt.Errorf("failed to create directory: %w", err)
		}
		if err := e.FS.Rename(src, op.DestPath); err != nil {
			return false, fmt.Errorf("failed to move file: %w", err)
		}

		e.Operations = append(e.Operations, op)
		return true, nil

	case core.OpDelete:
		if err := e.FS.Remove(src); err != nil {
			return false, fmt.Errorf("failed to delete file: %w", err)
		}
		return true, nil
	}

	return false, nil
}
