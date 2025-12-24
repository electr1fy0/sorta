package internal

import (
	"fmt"
	"os"
	"path/filepath"
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
			matchers[i] = Matcher{Raw: strings.TrimSpace(k)}
		}
		confData.Matchers = append(confData.Matchers, matchers)

		return &ConfigSorter{&confData}, nil
	}

	var confData *ConfigData

	sortaDir, err := GetSortaDir()
	if err != nil {
		return nil, err
	}
	defaultPath := filepath.Join(sortaDir, "config")
	var localPath string
	if configPath == defaultPath {
		localPath = filepath.Join(folderPath, ".sorta", "config")
	}

	_, err = os.Open(localPath)
	if err != nil {
		if os.IsNotExist(err) {
			confData, err = ParseConfig(configPath)
		} else {
			return nil, err
		}
	} else {
		confData, err = ParseConfig(localPath)
	}

	if err != nil {
		return nil, err
	}

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
