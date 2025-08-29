package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type ConfigData struct {
	foldernames []string
	keywords    [][]string
}

func main() {
	path := getPath()
	fmt.Println("Dir: ", path)
	filterFiles(path)
	// configData := parseConfig() // currently WIP

}

func createDefaultFolders(path string) {
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

		var isHidden bool = []rune(filename)[0] == '.'

		createDefaultFolders(path)
		if !isHidden {
			switch filepath.Ext(filename) {
			case ".pdf":
				moveFile(path, "docs", filename)
			case ".png", ".jpg", ".jpeg":
				moveFile(path, "images", filename)
			}
		}
	}
}

func readConfigFile() string {
	home, _ := os.UserHomeDir()
	configBytes, _ := os.ReadFile(filepath.Join(home, ".sorta-config"))
	config := string(configBytes)
	return config
}

func parseConfig() ConfigData {
	config := readConfigFile()
	var configData ConfigData

	lineCount := strings.Count(config, "\n")
	if []rune(config)[len(config)-1] != rune('\n') {
		lineCount++
	}
	configData.foldernames = make([]string, lineCount)
	configData.keywords = make([][]string, lineCount)
	for i := range lineCount {
		configData.keywords[i] = make([]string, 50)
	}

	i := 0
	for line := range strings.Lines(config) {

		input := strings.Split(line, ",")

		last := input[len(input)-1]
		last = strings.Trim(last, "\n ")

		lastSplit := strings.Split(last, "=")

		input[len(input)-1] = lastSplit[0]
		output := lastSplit[1]
		configData.foldernames[i] = output
		configData.keywords[i] = input
		i++
	}

	return configData
}

// do custom patterns read from a config file where if a file starts with these certain words then create those certain folders
// like fallsem, wintersem creates a vtop folder
//

// add duplicate removal

func createCustomFolder(path, foldername string) {
	os.MkdirAll(filepath.Join(path, foldername), 0700)

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
