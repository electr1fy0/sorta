package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// The logic:
// 1. If file ends with png, jpeg, mkdir images
// 2. rename all files to the dir
// 3. etc
// 4. rest of the files go to random

// Check if dir has images (img count != 0)
// Mkdir

func main() {
	path := getPath()
	fmt.Println("Dir: ", path)
	filterFiles(path)
}

func createFolders(path string) {
	os.MkdirAll(filepath.Join(path, "docs"), 0700)
	os.MkdirAll(filepath.Join(path, "images"), 0700)
}

func moveFile(folder, subfolder, filename string) {
	os.Rename(filepath.Join(folder, filename), filepath.Join(folder, subfolder, filename))
}

func filterFiles(path string) {
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Fatalln("Error joining path: ", err)
	}
	for i, entry := range entries {
		filename := entry.Name()
		fmt.Println(i+1, filename)

		createFolders(path)

		switch filepath.Ext(filename) {
		case ".pdf":
			moveFile(path, "docs", filename)
		case ".png", ".jpg", ".jpeg":
			moveFile(path, "images", filename)
		}

	}
}

func getPath() string {
	fmt.Println("Enter the directory (relative to ~/):")
	var dir string
	fmt.Scanf("%s", &dir)
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln("Error joining path", err)
	}
	path := filepath.Join(home, dir)
	return path
}
