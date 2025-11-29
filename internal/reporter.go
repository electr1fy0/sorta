package internal

import (
	"fmt"
	"path/filepath"
)

func (r *SortResult) PrintSummary() {
	fmt.Println("--------------------------------------------------")
	fmt.Printf("  %sMoved:%s   %d\n", ansiGreen, ansiReset, r.Moved)
	fmt.Printf("  %sSkipped:%s %d\n", ansiYellow, ansiReset, r.Skipped)
	if len(r.Errors) > 0 {
		fmt.Printf("  %sErrors:%s  %d\n", ansiRed, ansiReset, len(r.Errors))
		for _, err := range r.Errors {
			fmt.Printf("    - %v\n", err)
		}
	}
	fmt.Println("--------------------------------------------------")
}

func (r *Reporter) Report(op FileOperation, err error) {
	var tag string

	if err != nil {
		tag = ansiRed + "[ERR]" + ansiReset
		fmt.Printf("%s %s: %v\n", tag, op.Filename, err)
		return
	}

	if r.DryRun {
		tag = ansiYellow + "[DRY]" + ansiReset
	} else {
		tag = ansiGreen + "[OK] " + ansiReset
	}

	switch op.Type {
	case OpMove:
		srcDir := filepath.Dir(op.SourcePath)
		destDir := filepath.Dir(op.DestPath)

		if srcDir == destDir {
			newFilename := filepath.Base(op.DestPath)
			fmt.Printf("%s %s -> %s (%s)\n", tag, op.Filename, newFilename, humanReadable(op.Size))
		} else {
			destDirName := filepath.Base(destDir)
			fmt.Printf("%s %s -> %s/ (%s)\n", tag, op.Filename, destDirName, humanReadable(op.Size))
		}
	case OpDelete:
		if r.DryRun {
			tag = ansiYellow + "[DEL]" + ansiReset
		} else {
			tag = ansiRed + "[DEL]" + ansiReset
		}
		fmt.Printf("%s %s (%s)\n", tag, op.Filename, humanReadable(op.Size))
	}
}
