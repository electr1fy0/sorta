package ops

import (
	"fmt"
	"path/filepath"

	"github.com/electr1fy0/sorta/internal/core"
)

const (
	ansiReset = "\x1b[0m"
	ansiRed   = "\x1b[31m"
	ansiGreen = "\x1b[32m"
)

type Reporter struct{}

func (r *Reporter) Report(op core.FileOperation, err error) {
	var tag string

	if err != nil {
		tag = ansiRed + "[ERR]" + ansiReset
		return
	}

	tag = ansiGreen + "[OK] " + ansiReset

	switch op.OpType {
	case core.OpMove, core.OpRename, core.OpDedupe:
		srcDir := filepath.Dir(op.File.SourcePath)
		destDir := filepath.Dir(op.DestPath)

		if srcDir == destDir {
			newFilename := filepath.Base(op.DestPath)
			fmt.Printf("%s %s -> %s (%s)\n", tag, filepath.Base(op.File.SourcePath), newFilename, core.HumanReadable(op.Size))
		} else {
			destDirName := filepath.Base(destDir)
			fmt.Printf("%s %s -> %s/ (%s)\n", tag, filepath.Base(op.File.SourcePath), destDirName, core.HumanReadable(op.Size))
		}
	case core.OpDelete:
		tag = ansiRed + "[DEL]" + ansiReset
		fmt.Printf("%s %s (%s)\n", tag, filepath.Base(op.File.SourcePath), core.HumanReadable(op.Size))
	}
}
