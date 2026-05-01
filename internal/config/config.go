package config

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/electr1fy0/sorta/internal/core"
	"github.com/electr1fy0/sorta/templates"
)

type ConfigData struct {
	Foldernames []string
	Matchers    [][]Matcher
	Blacklist   []string
	Warnings    []string
}

type RuleSpec struct {
	Folder   string
	Keywords []string
}

type Matcher struct {
	Raw   string
	Regex *regexp.Regexp
}

func LoadConfig(explicitPath, targetDir string) (*ConfigData, string, error) {
	path, err := ResolveConfigPath(explicitPath, targetDir)
	if err != nil {
		return nil, "", err
	}

	cfg, err := ParseConfig(path)
	if err != nil {
		return nil, path, err
	}
	for _, warning := range cfg.Warnings {
		fmt.Fprintf(os.Stderr, "config warning: %s\n", warning)
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

	globalDir, err := core.GetSortaDir()
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
	return os.WriteFile(path, []byte(templates.DefaultConfig), 0644)
}

func parseMatcher(k string) (Matcher, error) {
	if trimmed, ok := strings.CutPrefix(k, "regex("); ok {
		if trimmed, ok = strings.CutSuffix(trimmed, ")"); ok {
			re, err := regexp.Compile(trimmed)
			if err != nil {
				return Matcher{}, fmt.Errorf("invalid regex %q: %w", trimmed, err)
			}
			return Matcher{Regex: re}, nil
		}
		return Matcher{}, fmt.Errorf("malformed regex matcher %q", k)
	}
	return Matcher{Raw: k}, nil
}

func ParseInline(s string) (*ConfigData, error) {
	normalized := strings.ReplaceAll(s, `\n`, "\n")
	return parseConfigFromReader(strings.NewReader(normalized))
}

func BuildConfigData(rules []RuleSpec) (*ConfigData, error) {
	configData := &ConfigData{
		Foldernames: make([]string, 0, len(rules)),
		Matchers:    make([][]Matcher, 0, len(rules)),
		Blacklist:   make([]string, 0),
		Warnings:    make([]string, 0),
	}

	for _, rule := range rules {
		folder := strings.TrimSpace(rule.Folder)
		if folder == "" {
			return nil, fmt.Errorf("rule folder cannot be empty")
		}
		matchers := make([]Matcher, 0, len(rule.Keywords))
		for _, keyword := range rule.Keywords {
			keyword = strings.TrimSpace(keyword)
			if keyword == "" {
				continue
			}
			matcher, err := parseMatcher(keyword)
			if err != nil {
				return nil, err
			}
			matchers = append(matchers, matcher)
		}
		if len(matchers) == 0 {
			return nil, fmt.Errorf("rule %q must include at least one keyword", folder)
		}
		configData.Foldernames = append(configData.Foldernames, folder)
		configData.Matchers = append(configData.Matchers, matchers)
	}

	if len(configData.Foldernames) == 0 {
		return nil, fmt.Errorf("at least one rule is required")
	}

	return configData, nil
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
	return parseConfigFromReader(file)
}

func parseConfigFromReader(r io.Reader) (*ConfigData, error) {
	var configData ConfigData
	configData.Foldernames = make([]string, 0, 50)
	configData.Matchers = make([][]Matcher, 0, 50)
	configData.Blacklist = make([]string, 0, 10)
	configData.Warnings = make([]string, 0, 8)

	scanner := bufio.NewScanner(r)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := scanner.Text()
		if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}

		if cleanedLine, found := strings.CutPrefix(line, "!"); found {
			configData.Blacklist = append(configData.Blacklist, strings.TrimSpace(cleanedLine))
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			configData.Warnings = append(configData.Warnings, fmt.Sprintf("line %d ignored: missing '='", lineNo))
			continue
		}

		folder := strings.TrimSpace(parts[0])
		if folder == "" {
			configData.Warnings = append(configData.Warnings, fmt.Sprintf("line %d ignored: empty folder name", lineNo))
			continue
		}
		keywords := strings.Split(parts[1], ",")
		matchers := make([]Matcher, 0, len(keywords))
		for _, k := range keywords {
			k = strings.TrimSpace(k)
			if k == "" {
				configData.Warnings = append(configData.Warnings, fmt.Sprintf("line %d: empty matcher in folder %q", lineNo, folder))
				continue
			}
			m, err := parseMatcher(k)
			if err != nil {
				configData.Warnings = append(configData.Warnings, fmt.Sprintf("line %d: %v", lineNo, err))
				continue
			}
			matchers = append(matchers, m)
		}
		configData.Foldernames = append(configData.Foldernames, folder)
		configData.Matchers = append(configData.Matchers, matchers)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	if len(configData.Foldernames) == 0 {
		return nil, fmt.Errorf("config is empty. Add keywords to .sorta-config in home directory")
	}

	return &configData, nil
}

func Categorize(configData ConfigData, filename string) string {
	fallback := ""
	lowerFilename := strings.ToLower(filename)
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
			if strings.Contains(lowerFilename, strings.ToLower(matcher.Raw)) {
				return foldername
			}
		}
	}

	return fallback
}
