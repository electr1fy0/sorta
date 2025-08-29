package main

import (
	"os"
	"path/filepath"
)

func main() {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".sorta-config")

	os.WriteFile(path, []byte("you=meow\n\"she=her"), 0600)
}
