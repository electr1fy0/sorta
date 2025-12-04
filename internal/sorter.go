package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
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

func logHistory(t Transaction) {
	data, _ := json.Marshal(t)
	data = append(data, '\n')

	home, _ := os.UserHomeDir()
	logPath := filepath.Join(home, ".sorta", "history.log")
	f, _ := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	f.Write(data)
	LogCnt++
	defer f.Close()
}

var transaction Transaction

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

	transaction.ID = time.Now().String()
	transaction.Root = filePaths[0].BaseDir
	transaction.Operations = ops
	logHistory(transaction)
	return ops, nil
}

func readHistory(root string) (Transaction, error) {
	home, err := os.UserHomeDir()
	historyPath := filepath.Join(home, ".sorta", "history.log")

	data, err := os.ReadFile(historyPath)
	var undoT Transaction
	lines := strings.Split(string(data), "\n")

	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		err = json.Unmarshal([]byte(line), &undoT)
		if undoT.Root == root {
			if undoT.Type == TUndo {
				return Transaction{}, fmt.Errorf("last operation in %s was already undone", root)
			}

			return undoT, err
		}
	}
	return Transaction{}, err
}

type TransactionType int

const (
	TAction TransactionType = iota
	TUndo
)

func Undo(path string) error {
	if !filepath.IsAbs(path) {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		path = filepath.Join(home, path)
	}
	t, err := readHistory(path)
	t.Type = TUndo
	logHistory(t)

	var executor Executor
	for _, op := range t.Operations {

		op.SourcePath, op.DestPath = op.DestPath, op.SourcePath
		executor.Execute(op)
	}
	return err
}

func (s *ConfigSorter) GetBlacklist() []string {
	return s.configData.Blacklist
}
