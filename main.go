package main

import (
	"github.com/electr1fy0/sorta/cmd"
)

// TODO:
// Blacklist foldernames
// make logging tree like
// cache checksum
// docs:
// - file 1
// - file 2
// images:
// - img 1
// - img 2
// sort by size by calculating average size
// sort by data
// more expressive summary
// concurrency in system calls
// regex support in config
// blacklist / whitelist (like gitignore)
// interactive mode: ask users what to do with unmatched files
// something called MIME type. use that instead of ext

func main() {
	cmd.Execute()
}
