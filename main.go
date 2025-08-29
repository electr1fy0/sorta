package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// type ConfigData struct {
// 	foldernames []string
// 	keywords    [][]string
// }

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
			switch filepath.Ext(filename) {
			case ".pdf":
				err := moveFile(path, "docs", filename)
				if err != nil {
					return err
				}
			case ".png", ".jpg", ".jpeg":
				err := moveFile(path, "images", filename)
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
				createFolder(path, foldername)
			}
			moveFile(path, foldername, filename)
		}
	}
	return nil
}

func readConfigFile() (string, error) {
	home, _ := os.UserHomeDir()
	configBytes, err := os.ReadFile(filepath.Join(home, ".sorta-config"))
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

	i := 0
	for line := range strings.Lines(config) {

		input := strings.Split(line, ",")
		last := input[len(input)-1]
		last = strings.Trim(last, "\n ")

		lastSplit := strings.Split(last, "=")

		input[len(input)-1] = lastSplit[0]
		output := lastSplit[1]
		configData[output] = input
		i++
	}

	return configData, nil
}

// todo: add duplicate removal

func getPathAndMode() (string, int) {
	fmt.Println("Enter the directory (relative to home dir):")
	fmt.Print("~/")
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
