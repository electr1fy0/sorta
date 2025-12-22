package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	logHistory(transaction)
	return ops, nil
}

// func readHistory(root string) (Transaction, error) {
// 	home, err := os.UserHomeDir()
// 	historyPath := filepath.Join(home, ".sorta", "history.log")

// 	data, err := os.ReadFile(historyPath)
// 	var undoT Transaction
// 	lines := strings.Split(string(data), "\n")

// 	for i := len(lines) - 1; i >= 0; i-- {
// 		line := lines[i]
// 		err = json.Unmarshal([]byte(line), &undoT)
// 		if undoT.Root == root {
// 			if undoT.Type == TUndo {
// 				return Transaction{}, fmt.Errorf("last operation in %s was already undone", root)
// 			}

// 			return undoT, err
// 		}
// 	}
// 	return Transaction{}, err
// }

type TransactionType int

const (
	TAction TransactionType = iota
	TUndo
)

// func Undo(path string) error {
// 	if !filepath.IsAbs(path) {
// 		home, err := os.UserHomeDir()
// 		if err != nil {
// 			return err
// 		}
// 		path = filepath.Join(home, path)
// 	}
// 	t, err := readHistory(path)
// 	t.Type = TUndo
// 	logHistory(t)

// 	var executor Executor
// 	for _, op := range t.Operations {

// 		op.SourcePath, op.DestPath = op.DestPath, op.SourcePath
// 		executor.Execute(op)
// 	}
// 	return err
// }

func (s *ConfigSorter) GetBlacklist() []string {
	return s.configData.Blacklist
}
