package ops

import (
	"context"
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
	return undoWithFS(path, core.OSFileSystem{})
}

func undoWithFS(path string, fs core.FileSystem) error {
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
		return fmt.Errorf("cannot undo irreversible operation (e.g. contains deletes or used --nuke)")
	}

	undoOps := make([]core.FileOperation, 0, len(t.Operations))
	for _, op := range t.Operations {
		undoOp := op
		undoOp.File.SourcePath, undoOp.DestPath = op.DestPath, op.File.SourcePath
		undoOps = append(undoOps, undoOp)
	}

	if _, err := applyOperationsWithoutHistory(context.TODO(), path, undoOps, &Executor{}, &Reporter{}, fs); err != nil {
		return fmt.Errorf("failed to undo operations: %w", err)
	}

	t.TType = core.TUndo
	if err := LogToHistory(t); err != nil {
		return fmt.Errorf("failed to log undo to history: %w", err)
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

	undoneIDs := make(map[string]bool)
	var allTransactions []core.Transaction

	// First pass: collect all transactions and identify undone ones
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var t core.Transaction
		if err := json.Unmarshal([]byte(line), &t); err != nil {
			continue
		}
		if len(t.Operations) > 0 && t.Operations[0].File.RootDir == root {
			allTransactions = append(allTransactions, t)
			if t.TType == core.TUndo {
				undoneIDs[t.ID] = true
			}
		}
	}

	// Second pass: find the last TAction that hasn't been undone
	for i := len(allTransactions) - 1; i >= 0; i-- {
		t := allTransactions[i]
		if t.TType == core.TUndo {
			return core.Transaction{}, fmt.Errorf("last operation in %s was already undone: %w", root, ErrAlreadyUndone)
		}
		if t.TType == core.TAction && !undoneIDs[t.ID] {
			return t, nil
		}
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
			continue
		}
		transactions = append(transactions, t)
	}

	return transactions, nil
}
