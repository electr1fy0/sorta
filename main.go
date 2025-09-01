package main

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type ConfigData map[string][]string

// TODO:
// scan directories recursively for duplicates
// concurrency in system calls
// config:
// - handle multi word keywords
// - add comment support with template explanation of config format

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
	var hashes map[string]string = make(map[string]string, len(entries))
	for _, entry := range entries {
		filename := entry.Name()

		var isHidden bool = []rune(filename)[0] == '.'
		var isDir = entry.IsDir()
		if isDir {
			filename += " (dir)"
		}

		if isHidden || isDir {
			continue
		}

		switch sortMode {
		case 0:
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
		case 1:
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
			}
			err = moveFile(path, foldername, filename)
			if err != nil {
				return err
			}
		case 2:
			fullpath := filepath.Join(path, filename)
			data, _ := os.ReadFile(fullpath)
			checksum256 := sha256.Sum256(data)
			digest := fmt.Sprintf("%x", checksum256)
			fmt.Println(digest)
			if _, exists := hashes[digest]; exists {
				os.MkdirAll(filepath.Join(path, "duplicates"), 0700)
				err := os.Rename(fullpath, filepath.Join(path, "duplicates", filename))
				if err != nil {
					return err
				}
			} else {
				hashes[digest] = fullpath
			}
		}

		// fmt.Printf("%-5d   | %s\n", i+1, filename)

	}
	fmt.Println("Files sorted successfuly")
	return nil
}

func createConfig() error {
	home, err := os.UserHomeDir()
	path := filepath.Join(home, ".sorta-config")

	_, err = os.Create(path)
	if err != nil {
		return err
	}
	return nil
}

func readConfigFile() (string, error) {
	home, _ := os.UserHomeDir()
	configName := ".sorta-config"
	configBytes, err := os.ReadFile(filepath.Join(home, configName))
	if err != nil {
		err = createConfig()
		if err != nil {
			return "", err
		}
		fmt.Println("Config file is empty. Add keywords to .sorta-config in home directory") // i'm lying here
		os.Exit(1)
	}

	config := string(configBytes)
	if strings.TrimSpace(config) == "" {
		fmt.Println("Config file is empty. Add keywords to .sorta-config in home directory")
		os.Exit(1)
	}
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
		if strings.HasPrefix(line, "//") {
			continue
		}
		input := strings.Split(line, ",")
		last := strings.TrimSpace(input[len(input)-1])

		lastSplit := strings.Split(last, "=")

		input[len(input)-1] = lastSplit[0]
		output := lastSplit[1]
		configData[output] = input
	}

	return configData, nil
}

func getPathAndMode() (string, int) {
	fmt.Println("Enter the directory (relative to home dir, no quotes):")
	fmt.Print("~/")
	var dir string
	var mode int

	reader := bufio.NewReader(os.Stdin)
	dir, _ = reader.ReadString('\n')
	dir = strings.TrimSpace(dir)
	fmt.Println("Choose mode index:")
	fmt.Println("0: Sort based on file extension")
	fmt.Println("1: Sort based on keywords in config")
	fmt.Println("2: Filter duplicates")

	fmt.Fscanln(reader, &mode)
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln("Error joining path", err)
	}
	path := filepath.Join(home, dir)
	return path, mode
}
