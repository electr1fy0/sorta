package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	home, _ := os.UserHomeDir()
	configName := ".sorta-config"

	path := filepath.Join(home, configName)
	file, _ := os.ReadFile(path)
	fmt.Print(string(file))
}
