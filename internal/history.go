package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func logToHistory(transaction Transaction) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	historyPath := filepath.Join(home, ".sorta", "history")
	f, err := os.OpenFile(historyPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	data, err := json.Marshal(transaction)

	_, err = f.Write([]byte(string(data) + "\n"))
	return err
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

	t.TType = TUndo
	logToHistory(t)
	var executor Executor
	for _, op := range t.Operations {
		op.File.SourcePath, op.DestPath = op.DestPath, op.File.SourcePath
		executor.Execute(op)
	}
	return nil
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
			if transaction.TType == TUndo {

				return Transaction{}, fmt.Errorf("last operation in %s was already undone", root)
			}

			return transaction, err
		}
	}
	return transaction, err
}
