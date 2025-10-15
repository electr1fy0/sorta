package main

import (
	"github.com/electr1fy0/sorta/cmd"
)

// TODO:
// move to trash folder instead of deleting
// a very interactive mode like
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
// Add * to add rest of the files to others * matches all

func main() {
	cmd.Execute()
}
