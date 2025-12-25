package internal

import (
	"fmt"
	"path/filepath"
)

func (r *SortResult) PrintSummary() {
	fmt.Println("--------------------------------------------------")
	if r.Moved > 0 {
		fmt.Printf("  %sMoved:%s %d\n", ansiGreen, ansiReset, r.Moved)
	} else if r.Deduped > 0 {
		fmt.Printf("  %sDeduped:%s %d\n", ansiGreen, ansiReset, r.Deduped)
	} else if r.Renamed > 0 {
		fmt.Printf("  %sRenamed:%s %d\n", ansiGreen, ansiReset, r.Renamed)
	}

	fmt.Printf("  %sDeleted:%s %d\n", ansiRed, ansiReset, r.Deleted)
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
		return
	}

	if r.DryRun {
		tag = ansiYellow + "[DRY]" + ansiReset
	} else {
		tag = ansiGreen + "[OK] " + ansiReset
	}

	switch op.OpType {
	case OpMove:
		srcDir := filepath.Dir(op.File.SourcePath)
		destDir := filepath.Dir(op.DestPath)

		if srcDir == destDir {
			newFilename := filepath.Base(op.DestPath)
			fmt.Printf("%s %s -> %s (%s)\n", tag, filepath.Base(op.File.SourcePath), newFilename, humanReadable(op.Size))
		} else {
			destDirName := filepath.Base(destDir)
			fmt.Printf("%s %s -> %s/ (%s)\n", tag, filepath.Base(op.File.SourcePath), destDirName, humanReadable(op.Size))
		}
	case OpDelete:
		if r.DryRun {
			tag = ansiYellow + "[DEL]" + ansiReset
		} else {
			tag = ansiRed + "[DEL]" + ansiReset
		}
		fmt.Printf("%s %s (%s)\n", tag, filepath.Base(op.File.SourcePath), humanReadable(op.Size))
	}
}
