package internal

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

func NewConfigSorter(folderPath, configPath, inline string) (*ConfigSorter, error) {

	inline = strings.TrimSpace(inline)

	if inline != "" {
		var confData ConfigData
		confData.Foldernames = make([]string, 0, 1)
		confData.Matchers = make([][]Matcher, 0, 1)

		parts := strings.Split(inline, "=")
		if len(parts) < 2 {
			return nil, fmt.Errorf("failed to parse inline syntax: %s", inline)
		}
		foldername := strings.TrimSpace(parts[0])
		keywords := strings.Split(parts[1], ",")

		confData.Foldernames = append(confData.Foldernames, foldername)

		matchers := make([]Matcher, len(keywords))
		for i, k := range keywords {
			k = strings.TrimSpace(k)
			if trimmed, ok := strings.CutPrefix(k, "regex("); ok {
				if trimmed, ok = strings.CutSuffix(trimmed, ")"); ok {
					fmt.Println(k)
					regex, err := regexp.Compile(trimmed)
					if err != nil {
						return nil, err
					}
					matchers[i] = Matcher{Regex: regex}
				}
			} else {
				matchers[i] = Matcher{Raw: strings.TrimSpace(k)}
			}
		}
		confData.Matchers = append(confData.Matchers, matchers)

		return &ConfigSorter{&confData}, nil
	}

	var confData *ConfigData

	var err error
	confData, _, err = LoadConfig(configPath, folderPath)
	if err != nil {
		return nil, err
	}
	// Log which config is being used if needed, or maybe just rely on LoadConfig
	// fmt.Printf("Using config: %s\n", loadedPath) // Optional debug

	return &ConfigSorter{
		configData: confData,
	}, nil
}

func (s *ConfigSorter) Decide(files []FileEntry) ([]FileOperation, error) {
	ops := make([]FileOperation, 0, 10)

	for _, file := range files {
		filename := filepath.Base(file.SourcePath)
		destFolder := categorize(*s.configData, filename)

		if destFolder == "" {
			ops = append(ops, FileOperation{OpType: OpSkip})
		} else {
			ops = append(ops, FileOperation{
				OpType:   OpMove,
				File:     file,
				Size:     file.Size,
				DestPath: filepath.Join(file.RootDir, destFolder, filename),
			})
		}
	}

	return ops, nil
}

func (s *ConfigSorter) GetBlacklist() []string {
	return s.configData.Blacklist
}
