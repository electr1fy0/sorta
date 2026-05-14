package ops

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/electr1fy0/sorta/internal/core"
)

func TestReadLastTransactionReturnsAlreadyUndone(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	root := filepath.Join(home, "workspace")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatalf("mkdir root: %v", err)
	}

	tx := core.Transaction{
		ID:    "tx-1",
		TType: core.TAction,
		Operations: []core.FileOperation{{
			OpType:   core.OpRename,
			File:     core.FileEntry{RootDir: root, SourcePath: filepath.Join(root, "a.txt")},
			DestPath: filepath.Join(root, "b.txt"),
		}},
	}
	undo := tx
	undo.TType = core.TUndo

	writeHistory(t, tx, undo)

	_, err := readLastTransaction(root)
	if !errors.Is(err, ErrAlreadyUndone) {
		t.Fatalf("expected ErrAlreadyUndone, got %v", err)
	}
}

func TestUndoRollbackDoesNotLogUndoOnFailure(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	root := filepath.Join(home, "workspace")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatalf("mkdir root: %v", err)
	}

	tx := core.Transaction{
		ID:    "tx-1",
		TType: core.TAction,
		Operations: []core.FileOperation{{
			OpType:   core.OpRename,
			File:     core.FileEntry{RootDir: root, SourcePath: filepath.Join(root, "a.txt")},
			DestPath: filepath.Join(root, "b.txt"),
		}},
	}
	writeHistory(t, tx)

	mfs := &mockFS{
		shouldFail: map[string]error{
			"rename": errors.New("disk full"),
		},
	}

	err := undoWithFS(root, mfs)
	if err == nil || !strings.Contains(err.Error(), "failed to undo operations") {
		t.Fatalf("expected undo failure, got %v", err)
	}

	history := readHistoryLines(t)
	if len(history) != 1 {
		t.Fatalf("expected only original history entry, got %d entries", len(history))
	}
	if history[0].TType != core.TAction {
		t.Fatalf("expected original action to remain untouched, got type %v", history[0].TType)
	}

	foundRollback := false
	for _, rename := range mfs.renames {
		if strings.Contains(rename[0], "staged") && rename[1] == filepath.Join(root, "b.txt") {
			foundRollback = true
			break
		}
	}
	if !foundRollback {
		t.Fatalf("expected failed undo to roll staged file back to current location")
	}
}

func writeHistory(t *testing.T, txs ...core.Transaction) {
	t.Helper()

	sortaDir, err := core.GetSortaDir()
	if err != nil {
		t.Fatalf("get sorta dir: %v", err)
	}
	if err := os.MkdirAll(sortaDir, 0755); err != nil {
		t.Fatalf("mkdir sorta dir: %v", err)
	}

	historyPath := filepath.Join(sortaDir, "history")
	var lines []string
	for _, tx := range txs {
		data, err := json.Marshal(tx)
		if err != nil {
			t.Fatalf("marshal tx: %v", err)
		}
		lines = append(lines, string(data))
	}
	if err := os.WriteFile(historyPath, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil {
		t.Fatalf("write history: %v", err)
	}
}

func readHistoryLines(t *testing.T) []core.Transaction {
	t.Helper()

	txs, err := GetHistory()
	if err != nil {
		t.Fatalf("get history: %v", err)
	}
	return txs
}
