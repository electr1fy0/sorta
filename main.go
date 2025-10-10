package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/electr1fy0/sorta/internal"
	"github.com/spf13/cobra"
)

const (
	ansiReset  = "\033[0m"
	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiCyan   = "\033[36m"
)

// TODO:
// fix recursive by appending subfolder path to dest or sth, idk
// move to trash folder instead of deleting
// a very interactive mode like
// [?] Move file "weirdfile.xyz" (12 MB)?
// [y] docs   [i] images   [m] movies   [s] skip
// make logging tree like
// cache checksum
// a flag to just nuke duplicates
// docs:
// - file 1
// - file 2
// images:
// - img 1
// - img 2
// sort by size by calculating average size
// sort by data
// more expressive summary
// scan directories recursively for duplicates
// concurrency in system calls
// regex support in config
// blacklist / whitelist (like gitignore)
// interactive mode: ask users what to do with unmatched files
// something called MIME type. use that instead of ext
// Add * to add rest of the files to others * matches all
// Undo all sort and bring to root

var state = &internal.State{}

var cmd = &cobra.Command{
	Short: "CLI to sort files based on extension and keywords",
	Use:   "sorta []",
	Run: func(cmd *cobra.Command, args []string) {
		for _, val := range args {
			state.CliDir += val + " "
		}
	},
}

func GetDirAndMode() (string, int) {
	reader := bufio.NewReader(os.Stdin)
	var dir string
	var mode int
	var err error

	if strings.TrimSpace(state.CliDir) != "" {
		dir = state.CliDir
	} else {
		fmt.Println(ansiCyan + "Enter the directory (relative to home dir, no quotes):" + ansiReset)
		fmt.Print(ansiCyan + "~/" + ansiReset)
		dir, _ = reader.ReadString('\n')
	}

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln(ansiRed+"Error reading home directory:"+ansiReset, err)
	}
	dir = strings.TrimSpace(dir)
	dir = filepath.Join(home, dir)

	fmt.Println("\n" + ansiCyan + "Choose mode index:" + ansiReset)
	fmt.Println(ansiYellow+"0:"+ansiReset, "Sort based on file extension")
	fmt.Println(ansiYellow+"1:"+ansiReset, "Sort based on keywords in config")
	fmt.Println(ansiYellow+"2:"+ansiReset, "Filter duplicates")

	fmt.Fscanln(reader, &mode)
	return dir, mode
}

func main() {
	path, mode := GetDirAndMode()

	fmt.Println("Dir:", path)

	var sorter internal.Sorter
	switch mode {
	case 0:
		sorter = internal.NewExtensionSorter()
	case 1:
		sorter, _ = internal.NewConfigSorter()
	case 2:
		sorter = internal.NewDuplicateFinder()
	}

	executor := &internal.Executor{DryRun: state.DryRun}
	reporter := &internal.Reporter{
		DryRun: state.DryRun,
	}
	res, err := internal.FilterFiles(path, sorter, executor, reporter)
	if err != nil {
		fmt.Println("Error filtering files: ", err)
		os.Exit(1)
	}
	if err := internal.TopLargestFiles(path, 5); err != nil {
		fmt.Println("Error finding largest files:", err)
		os.Exit(1)
	}
	res.Print()
}

func init() {
	cmd.Flags().BoolVar(&state.DryRun, "dry", false, "Do a dry run")

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
