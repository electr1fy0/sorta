package internal

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ParseConfig() (*ConfigData, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}

	configPath := filepath.Join(home, ".sorta-config")
	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := createConfig(configPath); err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("config file created at %s. please add keywords to it", configPath)
		}
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()
	var configData ConfigData
	configData.foldernames = make([]string, 0, 50)
	configData.keywords = make([][]string, 0, 50)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "//") || strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			continue // Invalid line format
		}

		folder := strings.TrimSpace(parts[1])
		keywords := strings.Split(parts[0], ",")
		for i, k := range keywords {
			keywords[i] = strings.TrimSpace(k)
		}
		configData.foldernames = append(configData.foldernames, folder)
		configData.keywords = append(configData.keywords, keywords)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if len(configData.foldernames) == 0 {
		return nil, fmt.Errorf("config file is empty. Add keywords to .sorta-config in home directory")
	}

	return &configData, nil
}

func createConfig(path string) error {
	content := []byte(`// Config file for 'sorta'
//
// Each line defines how files should be sorted.
// Format: key1,key2,key3 = folderName
//
// - key1, key2, key3, etc are keywords to match in file names.
// - folderName is the target folder for those files.
// - You can list one or many keywords before the '='.
// - Lines starting with '//' are comments and ignored.
// - Make sure no spaces exist between the keys and values
// - * as a keyword matches all filenames which don't contain the other keywords
// Example:
// invoice,bill,txt=Finance
// track,song=Music
// notes,book=Study
// *=others`)

	if err := os.WriteFile(path, content, 0600); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	return nil
}

func categorize(configData ConfigData, filename string) string {
	var hasStar bool
	var fallback string
	for i, foldername := range configData.foldernames {
		for _, keyword := range configData.keywords[i] {
			if keyword == "*" {
				hasStar = true
				fallback = foldername
			}
			if strings.Contains(filename, keyword) {
				return foldername
			}
		}
	}

	if hasStar {
		return fallback
	}
	return ""
}
