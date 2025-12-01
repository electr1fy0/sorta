package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func NewExtensionSorter() *ExtensionSorter {
	return &ExtensionSorter{
		categories: map[string][]string{
			"docs":   {".pdf", ".docx", ".pages", ".md", ".txts"},
			"images": {".png", ".jpg", ".jpeg", ".heic", ".heif"},
			"movies": {".mp4", ".mov"},
			"slides": {".pptx"},
		},
	}
}

func (s *ExtensionSorter) Sort(BaseDir, dir, filename string, size int64) (FileOperation, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	for folder, extensions := range s.categories {
		if slices.Contains(extensions, ext) {
			return FileOperation{
				Type:       OpMove,
				SourcePath: filepath.Join(dir, filename),
				DestPath:   filepath.Join(BaseDir, folder, filename),
				Filename:   filename,
				Size:       size,
			}, nil
		}
	}

	return FileOperation{Type: OpSkip}, nil
}

func NewConfigSorter(folderPath, configPath string) (*ConfigSorter, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error determining home directory: %w", err)
	}
	defaultPath := filepath.Join(home, ".sorta", "config")
	var localPath string
	if configPath == defaultPath {
		localPath = filepath.Join(folderPath, ".sorta", "config")
	}

	var confData *ConfigData
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

func (s *ConfigSorter) Sort(filePaths []FilePath) ([]FileOperation, error) {
	ops := make([]FileOperation, 0, 10)

	for _, filePath := range filePaths {
		srcPath := filepath.Join(filePath.FullDir, filePath.Filename)
		folder := categorize(*s.configData, filePath.Filename, filepath.Ext(srcPath))

		if folder == "" {
			ops = append(ops, FileOperation{Type: OpSkip})
		} else {
			ops = append(ops, FileOperation{
				Type:       OpMove,
				SourcePath: srcPath,
				DestPath:   filepath.Join(filePath.BaseDir, folder, filePath.Filename),
				Filename:   filePath.Filename,
				Size:       filePath.Size,
			})
		}
	}
	return ops, nil
}

func (s *ConfigSorter) GetBlacklist() []string {
	return s.configData.Blacklist
}