package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// todo: add duplicate removal

type ConfigData map[string][]string

func main() {
	path, mode := getPathAndMode()
	fmt.Println("Dir:", path)

	err := filterFiles(path, mode)
	if err != nil {
		fmt.Println("Error filtering files: ", err)
		os.Exit(1)
	}

}
func createFolder(path, foldername string) error {
	return os.MkdirAll(filepath.Join(path, foldername), 0700)
}

func moveFile(folder, subfolder, filename string) error {
	err := os.Rename(filepath.Join(folder, filename), filepath.Join(folder, subfolder, filename))
	return err
}

func categorize(configData ConfigData, filename string) string {
	for foldername, keywords := range configData {
		for _, keyword := range keywords {
			if strings.Contains(filename, keyword) {
				return foldername
			}
		}
	}
	return ""
}

func filterFiles(path string, sortMode int) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Fatalln("Error joining path: ", err)
	}
	for i, entry := range entries {
		filename := entry.Name()

		var isHidden bool = []rune(filename)[0] == '.'
		var isDir = entry.IsDir()
		if isDir {
			filename += " (dir)"
		}
		fmt.Printf("%-5d   | %s\n", i+1, filename)
		if isHidden || isDir {
			continue
		}
		if sortMode == 0 {
			createFolder(path, "docs")
			createFolder(path, "images")
			createFolder(path, "movies")
			switch strings.ToLower(filepath.Ext(filename)) {
			case ".pdf", ".docx", ".pages", ".md", ".txt":
				err := moveFile(path, "docs", filename)
				if err != nil {
					return err
				}
			case ".png", ".jpg", ".jpeg", ".heic", ".heif", ".webp":
				err := moveFile(path, "images", filename)
				if err != nil {
					return err
				}
			case ".mp4", ".mov":
				err := moveFile(path, "movies", filename)
				if err != nil {
					return err
				}
			}

		} else {
			configData, err := parseConfig()
			if err != nil {
				return err
			}
			foldername := categorize(configData, filename)
			if foldername != "" {
				err := createFolder(path, foldername)
				if err != nil {
					return err
				}
			} else {
				fmt.Println("Folder name is empty")
				os.Exit(1)
			}
			err = moveFile(path, foldername, filename)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func readConfigFile() (string, error) {
	home, _ := os.UserHomeDir()
	configName := ".sorta-config"
	configBytes, err := os.ReadFile(filepath.Join(home, configName))
	if err != nil {
		return "", err
	}
	config := string(configBytes)
	return config, nil
}

func parseConfig() (ConfigData, error) {
	var configData ConfigData
	config, err := readConfigFile()
	if err != nil {
		return configData, err
	}

	lineCount := strings.Count(config, "\n")
	if []rune(config)[len(config)-1] != rune('\n') {
		lineCount++
	}

	configData = make(map[string][]string)

	for line := range strings.Lines(config) {
		input := strings.Split(line, ",")
		last := input[len(input)-1]
		last = strings.Trim(last, "\n ")

		lastSplit := strings.Split(last, "=")

		input[len(input)-1] = lastSplit[0]
		output := lastSplit[1]
		configData[output] = input
	}

	return configData, nil
}

func getPathAndMode() (string, int) {
	fmt.Println("Enter the directory (relative to home dir):")
	fmt.Print("~/")
	var dir string
	fmt.Scanf("%s", &dir)
	fmt.Println("Enter mode of sorting (0: extension based, 1: keyword based):")
	fmt.Scanf("\n")
	var mode int
	fmt.Scanf("%d", &mode)
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln("Error joining path", err)
	}
	path := filepath.Join(home, dir)
	return path, mode
}
