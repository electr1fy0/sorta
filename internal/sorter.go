package internal

import (
	"crypto/sha256"
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

func (s *ExtensionSorter) Sort(baseDir, dir, filename string, size int64) (FileOperation, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	for folder, extensions := range s.categories {
		if slices.Contains(extensions, ext) {
			return FileOperation{
				Type:       OpMove,
				SourcePath: filepath.Join(dir, filename),
				DestPath:   filepath.Join(baseDir, folder, filename),
				Filename:   filename,
				Size:       size,
			}, nil
		}
	}

	return FileOperation{Type: OpSkip}, nil
}

func NewConfigSorter() (*ConfigSorter, error) {
	confData, err := ParseConfig()
	if err != nil {
		return nil, err
	}
	return &ConfigSorter{
		configData: confData,
	}, nil
}

func (s *ConfigSorter) Sort(baseDir, dir, filename string, size int64) (FileOperation, error) {
	folder := categorize(*s.configData, filename)

	if folder == "" {
		return FileOperation{Type: OpSkip}, nil
	}
	return FileOperation{
		Type:       OpMove,
		SourcePath: filepath.Join(dir, filename),
		DestPath:   filepath.Join(baseDir, folder, filename),
		Filename:   filename,
		Size:       size,
	}, nil
}

func NewDuplicateFinder() *DuplicateFinder {
	return &DuplicateFinder{
		hashes: make(map[string]string),
	}
}

func (d *DuplicateFinder) Sort(baseDir, dir, filename string, size int64) (FileOperation, error) {
	fullPath := filepath.Join(dir, filename)
	if fullPath == filepath.Join(baseDir, "duplicates", filename) {
		return FileOperation{
			Type: OpSkip,
		}, nil
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return FileOperation{Type: OpSkip}, err
	}

	checksum := fmt.Sprintf("%x", sha256.Sum256(data))
	if _, exists := d.hashes[checksum]; !exists {
		d.hashes[checksum] = fullPath
		return FileOperation{
			Type: OpSkip}, nil
	}

	return FileOperation{
		Type:       OpMove,
		SourcePath: fullPath,
		DestPath:   filepath.Join(baseDir, "duplicates", filename),
		Filename:   filename,
		Size:       size,
	}, nil
}
