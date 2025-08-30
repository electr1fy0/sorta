package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, "Downloads/test")
	f1, _ := os.ReadFile(filepath.Join(path, "dora1.jpg"))
	f2, _ := os.ReadFile(filepath.Join(path, "dora2.jpg"))
	f3, _ := os.ReadFile(filepath.Join(path, "dora3.jpg"))

	digest1 := sha256.Sum256(f1)
	digest2 := sha256.Sum256(f2)
	digest3 := sha256.Sum256(f3)

	fmt.Printf("%x\n", digest1)
	fmt.Printf("%x\n", digest2)
	fmt.Printf("%x\n", digest3)
}
