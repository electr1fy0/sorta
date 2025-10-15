package internal

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func (r *SortResult) PrintandAskUndo() {
	fmt.Println("Moved:", r.Moved)
	fmt.Println("Skipped:", r.Skipped)
	fmt.Println("Errors:", len(r.Errors))

	fmt.Println("[?] Undo? [y/n]")
	input := bufio.NewReader(os.Stdin)
	confirm, err := input.ReadString('\n')
	if err != nil {
		fmt.Println("Error taking undo input: ", err)
	}

	if strings.TrimSpace(confirm) == "y" {
		undoExec.Undo = true
		for _, op := range Operations {
			undoExec.Execute(op)
		}
		fmt.Println("Changes reverted.")
	}
}

var undoExec = Executor{DryRun: false, Interactive: false}

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
