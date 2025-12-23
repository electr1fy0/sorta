package internal

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func ParseConfig(configPath string) (*ConfigData, error) {
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
	configData.Foldernames = make([]string, 0, 50)
	configData.Keywords = make([][]string, 0, 50)
	configData.Blacklist = make([]string, 0, 10)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}

		if cleanedLine, found := strings.CutPrefix(line, "!"); found {
			configData.Blacklist = append(configData.Blacklist, strings.TrimSpace(cleanedLine))
			continue
		}
		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			continue
		}

		folder := strings.TrimSpace(parts[0])
		keywords := strings.Split(parts[1], ",")
		for i, k := range keywords {
			keywords[i] = strings.TrimSpace(k)
		}
		configData.Foldernames = append(configData.Foldernames, folder)
		configData.Keywords = append(configData.Keywords, keywords)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if len(configData.Foldernames) == 0 {
		return nil, fmt.Errorf("config file is empty. Add keywords to .sorta-config in home directory")
	}

	return &configData, nil

}

func createConfig(path string) error {
	content := []byte(`// Config file for 'sorta'
// Config version: v0.4
//
// Each line defines how files should be sorted.
// Format: folderName = key1,key2,key3
//
// - folderName is the target folder for those files.
// - key1, key2, key3, etc are keywords to match in file names.
// - You can list one or many keywords after the '='.
// - Lines starting with '//' are comments and ignored.
// - Add a ! followed by a foldername to blacklist the folder from being touched by sorta
// - * as a keyword matches all filenames which don't contain the other keywords
// - . as a foldernames means the root folder that you passed to sorta.
// - To flatten the subfolder tree, use . = *
// - Use regex for kewyords. Wrap your expression with: regex()
// - foldername can also be a relative folderpath. e.g. foo/bar/oof = rab creates a folder tree.
//
// Example:
//
// Finance=invoice,bill,txt
// Music=track,song
// Study=notes,book
// 2024-Papers=regex(^PAP.*2024$)
// others=*
//
// Important folder that sorta won't move from:
// !my-secret-folder`)

	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	return nil
}

func matchKeyword(keyword, filename string) (bool, error) {
	if trimmed, found := strings.CutPrefix(keyword, "regex("); found {
		if trimmed, found := strings.CutSuffix(trimmed, ")"); found {
			return regexp.MatchString(trimmed, filename)
		} else {
			return false, fmt.Errorf("invalid keyword, regex bracket left open")
		}
	} else if strings.Contains(filename, keyword) {
		return true, nil
	}
	return false, nil
}

func categorize(configData ConfigData, filename string) string {
	fallback := ""
	for i, foldername := range configData.Foldernames {
		for _, keyword := range configData.Keywords[i] {
			if keyword == "*" {
				fallback = foldername
			}
			match, _ := matchKeyword(keyword, filename)
			if match {
				return foldername
			}
		}
	}

	return fallback
}
