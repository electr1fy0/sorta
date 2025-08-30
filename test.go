package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, "Downloads/test")

	entries, _ := os.ReadDir(path)

	for i, entry := range entries {
		info, _ := entry.Info()
		data := info.Sys()
		fmt.Println(i+1, info.Sys())
	}

}
