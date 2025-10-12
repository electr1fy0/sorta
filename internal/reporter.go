package internal

import "fmt"

func (r *SortResult) Print() {
	fmt.Println("Moved:", r.Moved)
	fmt.Println("Skipped:", r.Skipped)
	fmt.Println("Errors:", len(r.Errors))
}

func (r *Reporter) Report(op FileOperation, err error) {
	prefix := ansiGreen + "[OK]" + ansiReset
	if r.DryRun {
		prefix = ansiYellow + "[DRY]" + ansiReset
	}

	if err != nil {
		prefix = ansiRed + "[ERR]" + ansiReset
	}

	if op.Type == OpMove {
		fmt.Printf("%s %s -> %s (%s)\n", prefix, op.Filename, op.DestPath, humanReadable(op.Size))
	}
}
