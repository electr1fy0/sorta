package main

import (
	"fmt"
	"os"

	"github.com/electr1fy0/sorta/internal"
)

// TODO:
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

func main() {
	path, mode := internal.GetDirAndMode()
	fmt.Println("Dir:", path)

	err := internal.FilterFiles(path, mode)
	if err != nil {
		fmt.Println("Error filtering files: ", err)
		os.Exit(1)
	}
	if err := internal.TopLargestFiles(path, 5); err != nil {
		fmt.Println("Error finding largest files:", err)
	}
}
