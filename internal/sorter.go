package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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

var transaction Transaction

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

	// transaction.ID = time.Now().String()
	// transaction.Root = filePaths[0].BaseDir
	// transaction.Operations = ops
	return ops, nil
}

func readLastTransaction(root string) (Transaction, error) {
	home, err := os.UserHomeDir()
	historyPath := filepath.Join(home, ".sorta", "history")
	var transaction Transaction

	data, err := os.ReadFile(historyPath)
	if err != nil {
		return transaction, err
	}
	lines := strings.Split(string(data), "\n")

	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			continue
		}
		err = json.Unmarshal([]byte(line), &transaction)
		if err != nil {
			return transaction, err
		}

		if len(transaction.Operations) > 0 && transaction.Operations[0].File.RootDir == root {
			if transaction.Type == TUndo {
				// fmt.Println(transaction.Type)
				return Transaction{}, fmt.Errorf("last operation in %s was already undone", root)
			}

			return transaction, err
		}
	}
	return transaction, err
}

func Undo(path string) error {
	if !filepath.IsAbs(path) {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		path = filepath.Join(home, path)
	}
	t, err := readLastTransaction(path)
	if err != nil {
		return err
	}
	t.Type = TUndo

	var executor Executor
	for _, op := range t.Operations {
		op.File.SourcePath, op.DestPath = op.DestPath, op.File.SourcePath
		executor.Execute(op)
	}
	return nil
}

func (s *ConfigSorter) GetBlacklist() []string {
	return s.configData.Blacklist
}
