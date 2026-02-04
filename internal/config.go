package internal

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/electr1fy0/sorta/templates"
)

func LoadConfig(explicitPath, targetDir string) (*ConfigData, string, error) {
	path, err := ResolveConfigPath(explicitPath, targetDir)
	if err != nil {
		return nil, "", err
	}

	cfg, err := ParseConfig(path)
	if err != nil {
		return nil, path, err
	}
	return cfg, path, nil
}

func ResolveConfigPath(explicitPath, targetDir string) (string, error) {
	if explicitPath != "" {
		return explicitPath, nil
	}

	if targetDir != "" {
		localPath := filepath.Join(targetDir, ".sorta", "config")
		if _, err := os.Stat(localPath); err == nil {
			return localPath, nil
		}
	}

	globalDir, err := GetSortaDir()
	if err != nil {
		return "", err
	}
	globalPath := filepath.Join(globalDir, "config")

	if _, err := os.Stat(globalPath); os.IsNotExist(err) {
		if err := createGlobalConfig(globalDir, globalPath); err != nil {
			return "", fmt.Errorf("failed to create global config: %w", err)
		}
	}

	return globalPath, nil
}

func createGlobalConfig(dir, path string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if err := os.WriteFile(path, []byte(templates.DefaultConfig), 0644); err != nil {
		return err
	}

	return nil
}

func ParseConfig(configPath string) (*ConfigData, error) {
	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found at %s", configPath)
		}
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()
	var configData ConfigData
	configData.Foldernames = make([]string, 0, 50)
	configData.Matchers = make([][]Matcher, 0, 50)
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
		matchers := make([]Matcher, len(keywords))
		for i, k := range keywords {
			k = strings.TrimSpace(k)
			if trimmed, found := strings.CutPrefix(k, "regex("); found {
				if trimmed, found := strings.CutSuffix(trimmed, ")"); found {
					re, err := regexp.Compile(trimmed)
					if err == nil {
						matchers[i] = Matcher{Regex: re}
						continue
					}
				}
			}
			matchers[i] = Matcher{Raw: k}
		}
		configData.Foldernames = append(configData.Foldernames, folder)
		configData.Matchers = append(configData.Matchers, matchers)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if len(configData.Foldernames) == 0 {
		return nil, fmt.Errorf("config file is empty. Add keywords to .sorta-config in home directory")
	}

	return &configData, nil
}

func categorize(configData ConfigData, filename string) string {
	fallback := ""
	for i, foldername := range configData.Foldernames {
		for _, matcher := range configData.Matchers[i] {
			if matcher.Regex != nil {
				if matcher.Regex.MatchString(filename) {
					return foldername
				}
				continue
			}

			if matcher.Raw == "*" {
				fallback = foldername
			}
			if strings.Contains(filename, matcher.Raw) {
				return foldername
			}
		}
	}

	return fallback
}
