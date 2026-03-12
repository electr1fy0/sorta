package ops

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/electr1fy0/sorta/internal/core"
)

var (
	ErrAlreadyUndone = errors.New("last operation already undone")
	ErrNoHistory     = errors.New("no recorded operation found for this directory")
)

func LogToHistory(transaction core.Transaction) error {
	sortaDir, err := core.GetSortaDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(sortaDir, 0755); err != nil {
		return err
	}
	historyPath := filepath.Join(sortaDir, "history")
	data, err := json.Marshal(transaction)
	if err != nil {
		return err
	}
	return core.AppendLineAtomic(historyPath, string(data), 0644)
}

func Undo(path string) error {
	if !filepath.IsAbs(path) {
		var err error
		path, err = filepath.Abs(path)
		if err != nil {
			return err
		}
	}
	t, err := readLastTransaction(path)

	if err != nil {
		return err
	}

	if t.Irreversible {
		return fmt.Errorf("cannot undo irreversible operation (e.g. used --nuke)")
	}

	t.TType = core.TUndo
	if err := LogToHistory(t); err != nil {
		return err
	}

	var executor Executor
	for _, op := range t.Operations {
		op.File.SourcePath, op.DestPath = op.DestPath, op.File.SourcePath
		executor.Execute(op)
	}
	return nil
}

func readLastTransaction(root string) (core.Transaction, error) {
	sortaDir, err := core.GetSortaDir()
	if err != nil {
		return core.Transaction{}, err
	}
	historyPath := filepath.Join(sortaDir, "history")

	data, err := os.ReadFile(historyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return core.Transaction{}, fmt.Errorf("%w: %s", ErrNoHistory, root)
		}
		return core.Transaction{}, err
	}
	lines := strings.Split(string(data), "\n")

	var transaction core.Transaction
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			continue
		}
		if err := json.Unmarshal([]byte(line), &transaction); err != nil {
			return core.Transaction{}, err
		}

		if len(transaction.Operations) == 0 || transaction.Operations[0].File.RootDir != root {
			continue
		}
		if transaction.TType == core.TUndo {
			return core.Transaction{}, fmt.Errorf("last operation in %s was already undone: %w", root, ErrAlreadyUndone)
		}
		return transaction, nil
	}
	return core.Transaction{}, fmt.Errorf("%w: %s", ErrNoHistory, root)
}

func GetHistory() ([]core.Transaction, error) {
	sortaDir, err := core.GetSortaDir()
	if err != nil {
		return nil, err
	}
	historyPath := filepath.Join(sortaDir, "history")

	data, err := os.ReadFile(historyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var transactions []core.Transaction

	for line := range strings.Lines(string(data)) {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var t core.Transaction
		if err := json.Unmarshal([]byte(line), &t); err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}

	return transactions, nil
}
