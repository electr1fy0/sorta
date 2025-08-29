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
	path, mode := getPathAndMode()
	fmt.Println("Dir: ", path)

	filterFiles(path, mode)

}
func createCustomFolder(path, foldername string) {
	os.MkdirAll(filepath.Join(path, foldername), 0700)
}

func createDefaultFolders(path string) {
	os.MkdirAll(filepath.Join(path, "docs"), 0700)
	os.MkdirAll(filepath.Join(path, "images"), 0700)

}

func moveFile(folder, subfolder, filename string) {
	os.Rename(filepath.Join(folder, filename), filepath.Join(folder, subfolder, filename))
}

func categorize(configData ConfigData, filename string) string {
	for i, foldername := range configData.foldernames {
		for j := 0; j < len(configData.keywords[i]); j++ {
			if strings.Contains(filename, configData.keywords[i][j]) {
				return foldername
			}
		}
	}
	return ""
}

func filterFiles(path string, sortMode int) {
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Fatalln("Error joining path: ", err)
	}
	for i, entry := range entries {
		filename := entry.Name()
		fmt.Println(i+1, filename)

		var isHidden bool = []rune(filename)[0] == '.'
		if !isHidden {
			if sortMode == 0 {
				createDefaultFolders(path)
				switch filepath.Ext(filename) {
				case ".pdf":
					moveFile(path, "docs", filename)
				case ".png", ".jpg", ".jpeg":
					moveFile(path, "images", filename)
				}

			} else {
				configData := parseConfig()
				foldername := categorize(configData, filename)
				if foldername != "" {
					createCustomFolder(path, foldername)
				}
				moveFile(path, foldername, filename)
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

	i := 0
	for line := range strings.Lines(config) {

		input := strings.Split(line, ",")
		configData.keywords[i] = make([]string, len(input))

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

// todo: add duplicate removal

func getPathAndMode() (string, int) {
	fmt.Println("Enter the directory (relative to ~/):")
	var dir string
	fmt.Scanf("%s", &dir)
	fmt.Println("Enter mode of sorting (0: extension based, 1 : keyword based):")
	var mode int
	fmt.Scanf("%d", &mode)
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln("Error joining path", err)
	}
	path := filepath.Join(home, dir)
	return path, mode
}
