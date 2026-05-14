package ops

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/electr1fy0/sorta/internal/core"
)

type mockFS struct {
	renames    [][2]string
	removes    []string
	mkdirAlls  []string
	statCalls  []string
	shouldFail map[string]error
}

func (m *mockFS) Rename(oldpath, newpath string) error {
	if err, ok := m.shouldFail["rename"]; ok {
		// Only fail if we are in the commit phase (second rename)
		if strings.Contains(oldpath, "staged") && !strings.Contains(newpath, "staged") {
			delete(m.shouldFail, "rename") // Fail only once
			return err
		}
	}
	m.renames = append(m.renames, [2]string{oldpath, newpath})
	return nil
}

func (m *mockFS) Remove(path string) error {
	m.removes = append(m.removes, path)
	return nil
}

func (m *mockFS) RemoveAll(path string) error {
	return nil
}

func (m *mockFS) MkdirAll(path string, perm os.FileMode) error {
	m.mkdirAlls = append(m.mkdirAlls, path)
	return nil
}

func (m *mockFS) Stat(name string) (os.FileInfo, error) {
	m.statCalls = append(m.statCalls, name)
	return nil, nil // Return nil for simplicity in this basic mock
}

func (m *mockFS) ReadDir(name string) ([]os.DirEntry, error) {
	return nil, nil
}

func (m *mockFS) IsNotExist(err error) bool {
	return errors.Is(err, os.ErrNotExist)
}

func TestApplyOperationsSuccess(t *testing.T) {
	mfs := &mockFS{
		shouldFail: make(map[string]error),
	}

	executor := &Executor{FS: mfs}
	reporter := &Reporter{}

	ops := []core.FileOperation{
		{
			OpType: core.OpRename,
			File: core.FileEntry{SourcePath: "file1.txt"},
			DestPath: "file1_new.txt",
		},
	}

	_, err := ApplyOperationsCtxFS(context.Background(), ".", ops, executor, reporter, mfs)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	foundStaging := false
	for _, r := range mfs.renames {
		if r[0] == "file1.txt" && strings.Contains(r[1], "staged") {
			foundStaging = true
		}
	}
	if !foundStaging {
		t.Errorf("expected file1.txt to be staged")
	}

	foundCommit := false
	for _, r := range mfs.renames {
		if strings.Contains(r[0], "staged") && r[1] == "file1_new.txt" {
			foundCommit = true
		}
	}
	if !foundCommit {
		t.Errorf("expected staged file to be committed to file1_new.txt")
	}
}

func TestApplyOperationsRollback(t *testing.T) {
	mfs := &mockFS{
		shouldFail: map[string]error{
			"rename": errors.New("disk full"),
		},
	}

	executor := &Executor{FS: mfs}
	reporter := &Reporter{}

	ops := []core.FileOperation{
		{
			OpType: core.OpRename,
			File: core.FileEntry{SourcePath: "file1.txt"},
			DestPath: "file1_new.txt",
		},
	}

	// trigger failure only during the commit phase
	_, err := ApplyOperationsCtxFS(context.Background(), ".", ops, executor, reporter, mfs)
	if err == nil {
		t.Fatal("expected failure during commit phase, got success")
	}

	// verify rollback: staged file should be moved back to source
	foundRollback := false
	for _, r := range mfs.renames {
		if strings.Contains(r[0], "staged") && r[1] == "file1.txt" {
			foundRollback = true
		}
	}
	if !foundRollback {
		t.Errorf("expected staged file to be rolled back to file1.txt")
	}
}
